package mysqlrepo

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"ux-securities/backend/internal/migrations"
)

func TestDirectoryListsBeyondFiftyAndAllowsMissingQuote(t *testing.T) {
	db := openIngestionTestDB(t)
	resetIngestionTestSchema(t, db)
	if err := migrations.Apply(context.Background(), db, filepath.Join("..", "..", "..", "..", "..", "db", "migrations")); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 61; i++ {
		if _, err := db.Exec("INSERT INTO stocks (symbol,name,data_status,updated_at) VALUES (?,?,'unavailable',UTC_TIMESTAMP(3)) ON DUPLICATE KEY UPDATE name=VALUES(name)", fmt.Sprintf("%05d", 20000+i), fmt.Sprintf("Directory Equity %d", i)); err != nil {
			t.Fatal(err)
		}
	}
	repo := NewMarketRepository(db)
	items, err := repo.Search(context.Background(), "Directory Equity")
	if err != nil || len(items) != 61 {
		t.Fatalf("directory entries=%d err=%v, want 61", len(items), err)
	}
	if items[0].Symbol != "20001" || items[60].Symbol != "20061" {
		t.Fatal("directory order must be stable")
	}
	quote, err := repo.Quote(context.Background(), "20001")
	if err != nil {
		t.Fatalf("NULL quote must remain readable: %v", err)
	}
	if quote.LastPrice != nil || quote.DataStatus != "unavailable" {
		t.Fatalf("directory row manufactured quote: %+v", quote)
	}
}
