package mysqlrepo

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	marketdata "ux-securities/backend/internal/integration/marketdata"
)

func TestRuntimeDiagnosticMarketProvider(t *testing.T) {
	dsn := os.Getenv("RUNTIME_DIAGNOSTIC_DSN")
	if dsn == "" {
		t.Skip("RUNTIME_DIAGNOSTIC_DSN is not set")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repository := NewMarketRepository(db)
	quote, err := repository.Quote(context.Background(), "00700")
	if err != nil {
		t.Fatalf("repository quote: %v", err)
	}
	if quote.Source != "akshare" || quote.DataStatus != "delayed" || quote.Volume != 1200100 {
		t.Fatalf("repository quote = %+v, want akshare delayed quote with integer volume", quote)
	}
	provider := marketdata.New(nil, repository, marketdata.NewMockProvider())
	got, err := provider.Quote(context.Background(), "00700")
	if err != nil {
		t.Fatalf("provider quote: %v", err)
	}
	if got.Source != "akshare" || got.DataStatus != "delayed" {
		t.Fatalf("provider quote = %+v, want akshare delayed quote", got)
	}
	t.Logf("repository=%+v provider=%+v", quote, got)
}
