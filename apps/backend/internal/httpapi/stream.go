package httpapi

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	market "ux-securities/backend/internal/domain/market"
)

type MarketStreamHub struct {
	mu      sync.RWMutex
	subs    map[chan market.MarketEvent]struct{}
	origins []string
}

func NewMarketStreamHub(frontendOrigin string) *MarketStreamHub {
	return &MarketStreamHub{subs: make(map[chan market.MarketEvent]struct{}), origins: originPatterns(frontendOrigin)}
}

func (h *MarketStreamHub) Subscribe(ctx context.Context) <-chan market.MarketEvent {
	channel := make(chan market.MarketEvent, 8)
	h.mu.Lock()
	h.subs[channel] = struct{}{}
	h.mu.Unlock()
	go func() {
		<-ctx.Done()
		h.mu.Lock()
		delete(h.subs, channel)
		h.mu.Unlock()
	}()
	return channel
}

func (h *MarketStreamHub) Publish(event market.MarketEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for subscriber := range h.subs {
		select {
		case subscriber <- event:
		default:
		}
	}
}

func (h *MarketStreamHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	connection, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: h.origins})
	if err != nil {
		return
	}
	defer connection.Close(websocket.StatusNormalClosure, "")
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	for event := range h.Subscribe(ctx) {
		if err := wsjson.Write(ctx, connection, event); err != nil {
			return
		}
	}
}

func originPatterns(raw string) []string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" {
		return nil
	}
	return []string{parsed.Scheme + "://" + parsed.Host}
}
