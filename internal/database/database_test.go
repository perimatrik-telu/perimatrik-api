package database

import (
	"testing"
	"time"

	driver "github.com/go-sql-driver/mysql"
	"perimatrik-api-go/internal/config"
)

func TestConnectionTLS(t *testing.T) {
	for _, tc := range []struct {
		name, host, mode string
		wantTLS          bool
		wantError        bool
	}{
		{name: "TiDB default", host: "gateway.aws-eu-central-1.prod.shared.aws.tidbcloud.com", wantTLS: true},
		{name: "TiDB auto", host: "gateway.tidbcloud.com", mode: "auto", wantTLS: true},
		{name: "case insensitive hostname", host: "Gateway.TiDBCloud.COM", mode: "auto", wantTLS: true},
		{name: "local default", host: "127.0.0.1"},
		{name: "other remote auto", host: "mysql.example.com", mode: "auto"},
		{name: "require TLS", host: "mysql.example.com", mode: "true", wantTLS: true},
		{name: "explicit disable", host: "gateway.tidbcloud.com", mode: "false"},
		{name: "unrelated suffix", host: "not-tidbcloud.com", mode: "auto"},
		{name: "invalid mode", host: "localhost", mode: "typo", wantError: true},
		{name: "reject unverified TLS", host: "gateway.tidbcloud.com", mode: "skip-verify", wantError: true},
		{name: "reject plaintext fallback", host: "gateway.tidbcloud.com", mode: "preferred", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loc, err := time.LoadLocation("Asia/Jakarta")
			if err != nil {
				t.Fatal(err)
			}
			cfg := &config.Config{Location: loc, DB: config.DBConfig{
				Host: tc.host, Port: "4000", User: "user", Password: "p@ss:&word", Name: "test", TLS: tc.mode,
			}}
			dsn, err := connectionDSN(cfg)
			if tc.wantError {
				if err == nil {
					t.Fatal("expected invalid TLS setting to fail before connecting")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			parsed, err := driver.ParseDSN(dsn)
			if err != nil {
				t.Fatal(err)
			}
			if (parsed.TLS != nil) != tc.wantTLS {
				t.Fatalf("TLS enabled = %v, want %v", parsed.TLS != nil, tc.wantTLS)
			}
			if parsed.TLS != nil && (parsed.TLS.InsecureSkipVerify || parsed.TLS.ServerName != tc.host) {
				t.Fatal("TLS must verify the certificate and database hostname")
			}
			if parsed.AllowFallbackToPlaintext {
				t.Fatal("TLS must not fall back to plaintext")
			}
			if parsed.User != cfg.DB.User || parsed.Passwd != cfg.DB.Password || parsed.DBName != cfg.DB.Name ||
				!parsed.ParseTime || parsed.Loc.String() != loc.String() || parsed.Params["charset"] != "utf8mb4" ||
				parsed.Collation != "utf8mb4_unicode_ci" {
				t.Fatal("connection options changed unexpectedly")
			}
		})
	}
}
