package mysqlrepo

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWatchlistQueriesConstrainActorAndDuplicateRows(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	sourceBytes, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "watchlist.go"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	for _, fragment := range []string{
		"WHERE w.user_id = ?",
		"AND s.symbol = ?",
		"ON DUPLICATE KEY UPDATE",
		"DELETE w FROM watchlists",
	} {
		if !strings.Contains(source, fragment) {
			t.Fatalf("watchlist repository missing required SQL fragment %q", fragment)
		}
	}
}
