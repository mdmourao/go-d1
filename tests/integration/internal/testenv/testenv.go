package testenv

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/joho/godotenv"

	_ "github.com/mdmourao/go-d1"
	god1 "github.com/mdmourao/go-d1"
)

const DSNEnv = "HOST"
const CloudflareAccessClientIDEnv = "CF_ACCESS_CLIENT_ID"
const CloudflareAccessClientSecretEnv = "CF_ACCESS_CLIENT_SECRET"

var loadDotEnvOnce sync.Once

func loadDotEnv() {
	loadDotEnvOnce.Do(func() {
		path, ok := findDotEnv()
		if !ok {
			return
		}
		_ = godotenv.Load(path)
	})
}

func findDotEnv() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for {
		candidate := filepath.Join(dir, ".env")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func DSN(t testing.TB) string {
	t.Helper()
	loadDotEnv()
	dsn := os.Getenv(DSNEnv)
	if dsn == "" {
		t.Fatalf("environment variable %s is not set (and not found in .env)", DSNEnv)
	}
	return dsn
}

func CloudflareAccessCredentials(t testing.TB) (string, string) {
	t.Helper()
	loadDotEnv()
	id := os.Getenv(CloudflareAccessClientIDEnv)
	secret := os.Getenv(CloudflareAccessClientSecretEnv)
	if id == "" || secret == "" {
		t.Fatalf("environment variables %s and %s must be set (and not found in .env)", CloudflareAccessClientIDEnv, CloudflareAccessClientSecretEnv)
	}
	return id, secret
}

func OpenDB(t testing.TB) *sql.DB {
	t.Helper()
	dsn := DSN(t)
	id, secret := CloudflareAccessCredentials(t)

	connector, err := god1.NewConnector(
		dsn,
		god1.WithCloudflareAccess(id, secret),
	)
	if err != nil {
		t.Fatalf("failed to create connector: %v", err)
	}

	db := sql.OpenDB(connector)

	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping: %v", err)
	}
	return db
}
