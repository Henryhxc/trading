package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	market "ux-securities/backend/internal/domain/market"
)

func TestMarketStreamHubPublishesCommittedQuoteEvents(t *testing.T) {
	hub := NewMarketStreamHub("http://localhost:3000")
	if hub == nil {
		t.Fatal("NewMarketStreamHub returned nil")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := hub.Subscribe(ctx)
	hub.Publish(market.MarketEvent{Type: "quote.updated", Symbol: "00700", Source: "akshare", DataStatus: market.DataStatusDelayed, AsOf: time.Now()})
	select {
	case event := <-events:
		if event.Symbol != "00700" || event.Type != "quote.updated" {
			t.Fatalf("event = %+v, want committed quote event", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for published event")
	}
}

func TestMarketStreamRequiresSessionAndWritesJSONEvents(t *testing.T) {
	hub := NewMarketStreamHub("http://localhost:3000")
	server := httptest.NewServer(NewWithRealtime(&testAuthService{}, false, nil, nil, nil, nil, hub))
	defer server.Close()
	wsURL := "ws" + server.URL[len("http"):] + "/api/v1/markets/stream"
	header := http.Header{"Cookie": []string{"ux_session=synthetic-session"}}
	connection, _, err := websocket.Dial(context.Background(), wsURL, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		t.Fatalf("websocket dial: %v", err)
	}
	defer connection.Close(websocket.StatusNormalClosure, "")
	want := market.MarketEvent{Type: "quote.updated", Symbol: "00700", Source: "akshare", DataStatus: market.DataStatusDelayed, AsOf: time.Now()}
	hub.Publish(want)
	var got market.MarketEvent
	if err := wsjson.Read(context.Background(), connection, &got); err != nil {
		t.Fatalf("read event: %v", err)
	}
	if got.Type != want.Type || got.Symbol != want.Symbol || got.Source != want.Source {
		t.Fatalf("event = %+v, want %+v", got, want)
	}
}
