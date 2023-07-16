package mediaserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/csnewman/cathode/shared"
	dsdm "github.com/csnewman/dyndirect/go"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

const NetworkStoreKey = "network-store"

type NetworkStore struct {
	Entries map[string]DSDMEntry `json:"entries"`
}

type DSDMState string

const (
	DSDMStateUninitialized DSDMState = "uninitialized"
	DSDMStateAllocated     DSDMState = "allocated"
	DSDMStateActive        DSDMState = "active"
)

type DSDMEntry struct {
	State       DSDMState        `json:"state"`
	ID          uuid.UUID        `json:"id"`
	Domain      string           `json:"domain"`
	Token       string           `json:"token"`
	Certificate []byte           `json:"cert"`
	PrivateKey  []byte           `json:"cert_private_key"` //nolint:tagliatelle
	IssueDate   time.Time        `json:"issue_date"`       //nolint:tagliatelle
	Parsed      *tls.Certificate `json:"-"`
}

type NetworkManager struct {
	logger *slog.Logger
	db     *shared.DB
	store  *NetworkStore
	certs  map[string]*tls.Certificate
}

func NewNetworkManager(logger *slog.Logger, db *shared.DB) (*NetworkManager, error) {
	var store *NetworkStore

	logger.Info("Loading network store")

	err := db.Transact(true, func(tx *shared.Tx) error {
		if err := tx.Get(NetworkStoreKey, &store); err != nil {
			return err
		}

		if store != nil {
			return nil
		}

		logger.Info("No network store found. Generating")

		store = &NetworkStore{
			Entries: map[string]DSDMEntry{},
		}

		store.Entries[dsdm.DynDirect] = DSDMEntry{
			State: DSDMStateUninitialized,
		}

		return tx.Set(NetworkStoreKey, &store)
	})
	if err != nil {
		return nil, err
	}

	return &NetworkManager{
		logger: logger,
		db:     db,
		store:  store,
	}, nil
}

func (m *NetworkManager) Refresh(ctx context.Context) {
	certs := map[string]*tls.Certificate{}
	entries := map[string]DSDMEntry{}

	for server, currentEntry := range m.store.Entries {
		m.logger.Debug("Checking dsdm server", "server", server, "state", currentEntry.State)

		newEntry, err := m.refreshDSDM(ctx, server, currentEntry)
		if err != nil {
			m.logger.Error("Failed to refresh", "err", err)
		}

		entries[server] = newEntry

		err = m.db.Transact(true, func(tx *shared.Tx) error {
			return tx.Set(NetworkStoreKey, &m.store)
		})
		if err != nil {
			m.logger.Error("Failed to save", "err", err)
		}

		if newEntry.State == DSDMStateActive && newEntry.Parsed != nil {
			certs[strings.ToLower(newEntry.Domain)] = newEntry.Parsed
		}
	}

	m.certs = certs
	m.store.Entries = entries

	m.logger.Info("Network refreshed")
}

func (m *NetworkManager) refreshDSDM(ctx context.Context, server string, entry DSDMEntry) (DSDMEntry, error) {
	dc, err := dsdm.New(server)
	if err != nil {
		return entry, err
	}

	if entry.State == DSDMStateUninitialized {
		m.logger.Debug("Requesting subdomain", "server", server)

		resp, err := dc.RequestSubdomain(ctx)
		if err != nil {
			return entry, err
		}

		entry.ID = resp.Id
		entry.Token = resp.Token
		entry.Domain = resp.Domain
		entry.State = DSDMStateAllocated

		m.logger.Debug("Subdomain allocated", "server", server, "id", resp.Id, "domain", resp.Domain)
	}

	if entry.State == DSDMStateAllocated {
		m.logger.Debug("Requesting certificate", "server", server)

		cert, err := dc.AcquireCertificate(ctx, dsdm.AcquireCertificateRequest{
			ID:     entry.ID,
			Domain: entry.Domain,
			Token:  entry.Token,
			// TODO: Support other providers
			Provider:   dsdm.ProviderZeroSSL,
			KeyType:    certcrypto.RSA4096,
			Timeout:    time.Second * 120,
			SilenceLog: true,
		})
		if err != nil {
			return entry, err
		}

		entry.IssueDate = time.Now()
		entry.Certificate = cert.Certificate
		entry.PrivateKey = cert.PrivateKey
		entry.State = DSDMStateActive

		m.logger.Debug("Certificate issued", "server", server)
	}

	if entry.State == DSDMStateActive {
		cert, err := tls.X509KeyPair(entry.Certificate, entry.PrivateKey)
		if err != nil {
			return entry, err
		}

		entry.Parsed = &cert
	}

	return entry, nil
}

func (m *NetworkManager) Run(_ context.Context) {
	for _, entry := range m.store.Entries {
		m.logger.Info("Address", "url", fmt.Sprintf("https://127-0-0-1-v4.%s:8443", entry.Domain))
	}

	m.createServer(8443)
}

func (m *NetworkManager) createServer(port int) {
	hs := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion:     tls.VersionTLS12,
			GetCertificate: m.resolveCertificate,
		},
	}

	err := hs.ListenAndServeTLS("", "")
	m.logger.Error("temp", "err", err)
}

func (m *NetworkManager) resolveCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	m.logger.Debug("Resolving certificate", "name", info.ServerName)
	parts := strings.SplitN(info.ServerName, ".", 2)

	if len(parts) != 2 {
		return nil, nil //nolint:nilnil
	}

	cert, ok := m.certs[strings.ToLower(parts[1])]
	if !ok {
		return nil, nil //nolint:nilnil
	}

	return cert, nil
}
