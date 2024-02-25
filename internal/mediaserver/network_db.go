package mediaserver

import (
	"context"

	"github.com/csnewman/cathode/internal/db"
)

func getDSDMServers(_ context.Context, tx db.RTx) ([]string, error) {
	var servers []string

	rows, err := tx.Query(`SELECT server FROM dsdm_servers`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var res string
		if err := rows.Scan(&res); err != nil {
			return nil, err
		}

		servers = append(servers, res)
	}

	return servers, nil
}

func insertDSDMServer(_ context.Context, tx db.WTx, server string) error {
	return tx.Exec(`INSERT OR REPLACE INTO dsdm_servers (server) VALUES ($1)`, server)
}

func cleanupDSDMCerts(_ context.Context, tx db.WTx) error {
	return tx.Exec(`DELETE FROM dsdm_certs WHERE expire_date < datetime()`)
}

type dsdmEntry struct {
	domain string
	server string
	cert   []byte
	priKey []byte
}

func getDSDMCerts(_ context.Context, tx db.RTx) ([]dsdmEntry, error) {
	var entries []dsdmEntry

	rows, err := tx.Query(`SELECT domain, server, cert, pri_key FROM dsdm_certs`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var res dsdmEntry
		if err := rows.Scan(&res.domain, &res.server, &res.cert, &res.priKey); err != nil {
			return nil, err
		}

		entries = append(entries, res)
	}

	return entries, nil
}

func insertDSDMCert(_ context.Context, tx db.WTx, entry dsdmEntry) error {
	return tx.Exec(`
		INSERT OR REPLACE INTO dsdm_certs
		(domain, server, cert, pri_key, issue_date, expire_date)
		VALUES ($1, $2, $3, $4, datetime(), datetime('now', '+30 days'))`,
		entry.domain,
		entry.server,
		entry.cert,
		entry.priKey,
	)
}
