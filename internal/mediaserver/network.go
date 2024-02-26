package mediaserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/csnewman/cathode/internal/db"
	dsdm "github.com/csnewman/dyndirect/go"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type NetworkManager struct {
	logger *slog.Logger
	db     *db.DB
	certs  map[string]*tls.Certificate
	router *chi.Mux
}

func NewNetworkManager(logger *slog.Logger, db *db.DB) (*NetworkManager, error) {
	m := &NetworkManager{
		logger: logger,
		db:     db,
	}

	m.router = chi.NewRouter()
	m.router.Use(m.loggerMiddleware)
	m.router.Use(m.recoverMiddleware)

	m.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		ExposedHeaders:   []string{},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1, err := newV1API(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create v1: %w")
	}

	m.router.Mount("/api/v1", v1.router)
	m.router.HandleFunc("/api/*", m.fallbackAPIRoute)
	m.router.HandleFunc("/api", m.fallbackAPIRoute)
	m.router.HandleFunc("/*", m.fallbackPageRoute)

	return m, nil
}

func (m *NetworkManager) Refresh(ctx context.Context) {
	if err := m.refreshInner(ctx); err != nil {
		m.logger.Error("Refresh failed", "err", err)
	}
}

func (m *NetworkManager) refreshInner(ctx context.Context) error {
	m.logger.Info("Refreshing network")

	servers, err := db.ReadWithData(ctx, m.db, getDSDMServers)
	if err != nil {
		return fmt.Errorf("failed to fetch servers: %w", err)
	}

	// Default to dyndirect if no servers are configured
	if len(servers) == 0 {
		servers = append(servers, dsdm.DynDirect)

		err := m.db.Write(ctx, func(ctx context.Context, tx db.WTx) error {
			return insertDSDMServer(ctx, tx, dsdm.DynDirect)
		})
		if err != nil {
			return fmt.Errorf("failed to insert default server: %w", err)
		}
	}

	if err := m.db.Write(ctx, cleanupDSDMCerts); err != nil {
		return fmt.Errorf("failed to cleanup old certs: %w", err)
	}

	entries, err := db.ReadWithData(ctx, m.db, getDSDMCerts)
	if err != nil {
		return fmt.Errorf("failed to fetch certs: %w", err)
	}

	seenServers := make(map[string]struct{})
	newCerts := make(map[string]*tls.Certificate)

	for _, entry := range entries {
		parsed, err := tls.X509KeyPair(entry.cert, entry.priKey)
		if err != nil {
			return fmt.Errorf("failed to parse cert: %w", err)
		}

		seenServers[entry.server] = struct{}{}
		newCerts[entry.domain] = &parsed
	}

	m.certs = newCerts

	for _, server := range servers {
		if _, ok := seenServers[server]; ok {
			continue
		}

		err := m.refreshDSDM(ctx, server)
		if err != nil {
			m.logger.Error("Failed to refresh DSDM server", "server", server, "err", err)
		}
	}

	m.logger.Info("Network refreshed")

	return nil
}

func (m *NetworkManager) refreshDSDM(ctx context.Context, server string) error {
	dc, err := dsdm.New(server)
	if err != nil {
		return err
	}

	m.logger.Info("Requesting subdomain", "server", server)

	resp, err := dc.RequestSubdomain(ctx)
	if err != nil {
		return fmt.Errorf("failed to get subdomain: %w", err)
	}

	m.logger.Info("Subdomain allocated", "server", server, "id", resp.Id, "domain", resp.Domain)

	m.logger.Info("Requesting certificate", "server", server)

	cert, err := dc.AcquireCertificate(ctx, dsdm.AcquireCertificateRequest{
		ID:     resp.Id,
		Domain: resp.Domain,
		Token:  resp.Token,
		// TODO: Support other providers
		Provider:   dsdm.ProviderZeroSSL,
		KeyType:    certcrypto.RSA4096,
		Timeout:    time.Second * 120,
		SilenceLog: true,
	})
	if err != nil {
		return fmt.Errorf("failed to acquire cert: %w", err)
	}

	m.logger.Debug("Certificate issued", "server", server)

	parsed, err := tls.X509KeyPair(cert.Certificate, cert.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to parse cert: %w", err)
	}

	err = m.db.Write(ctx, func(ctx context.Context, tx db.WTx) error {
		return insertDSDMCert(ctx, tx, dsdmEntry{
			domain: resp.Domain,
			server: server,
			cert:   cert.Certificate,
			priKey: cert.PrivateKey,
		})
	})
	if err != nil {
		return fmt.Errorf("failed to store: %w", err)
	}

	m.certs[resp.Domain] = &parsed

	return nil
}

func (m *NetworkManager) Run(_ context.Context) {
	for dom := range m.certs {
		m.logger.Info(
			"Address",
			"url", fmt.Sprintf("https://127-0-0-1-v4.%s:8443", dom),
		)
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
		Handler: m.router,
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

func (m *NetworkManager) loggerMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		t1 := time.Now()
		defer func() {
			m.logger.Debug(
				"Request",
				"path", r.RequestURI,
				"status", ww.Status(),
				"size", ww.BytesWritten(),
				"time", time.Since(t1),
			)
		}()

		next.ServeHTTP(ww, r)
	}

	return http.HandlerFunc(fn)
}

func (m *NetworkManager) recoverMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				//nolint:errorlint,goerr113
				if rvr == http.ErrAbortHandler {
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(rvr)
				}

				m.logger.Error(
					"Request panic",
					"err", rvr,
					"stack", debug.Stack(),
				)

				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(http.StatusBadRequest)
				}
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func (m *NetworkManager) fallbackAPIRoute(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte("invalid api version"))
}

func (m *NetworkManager) fallbackPageRoute(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Nothing to see here"))
}
