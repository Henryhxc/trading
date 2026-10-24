package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	domain "ux-securities/backend/internal/domain/auth"
	market "ux-securities/backend/internal/domain/market"
	watchlist "ux-securities/backend/internal/domain/watchlist"
	"ux-securities/backend/internal/service/auth"
)

const sessionCookie = "ux_session"

var requestIDPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]{8,64}$`)

type AuthService interface {
	Register(context.Context, domain.Registration) (domain.User, string, time.Time, error)
	Login(context.Context, domain.Credentials) (domain.User, string, time.Time, error)
	Authenticate(context.Context, string) (domain.Actor, error)
	Logout(context.Context, string) error
}

type server struct {
	auth         AuthService
	cookieSecure bool
	market       MarketAPI
	watchlist    WatchlistAPI
	ingestion    IngestionAPI
	verifier     WebhookVerification
	stream       http.Handler
	trading      TradingAPI
}
type actorKey struct{}
type requestIDKey struct{}

type apiError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func New(service AuthService, cookieSecure bool) http.Handler {
	s := &server{auth: service, cookieSecure: cookieSecure}
	return s.routes()
}

type MarketAPI interface {
	Search(context.Context, string) ([]market.Symbol, error)
	Quote(context.Context, string) (market.Quote, error)
	Candles(context.Context, string, time.Time, time.Time, int) ([]market.Candle, error)
}
type WatchlistAPI interface {
	List(context.Context, domain.Actor) ([]watchlist.Item, error)
	Add(context.Context, domain.Actor, string) (watchlist.Item, bool, error)
	Delete(context.Context, domain.Actor, string) error
	Reorder(context.Context, domain.Actor, []string) error
}

type IngestionAPI interface {
	Ingest(context.Context, market.IngestionBatch) (market.IngestionResult, error)
}

type WebhookVerification interface {
	Verify(timestamp, deliveryID, signature string, body []byte) error
}

func NewWithMarket(service AuthService, cookieSecure bool, markets MarketAPI, lists WatchlistAPI) http.Handler {
	s := &server{auth: service, cookieSecure: cookieSecure, market: markets, watchlist: lists}
	return s.routes()
}

func NewWithTrading(service AuthService, cookieSecure bool, markets MarketAPI, lists WatchlistAPI, trades TradingAPI) http.Handler {
	s := &server{auth: service, cookieSecure: cookieSecure, market: markets, watchlist: lists, trading: trades}
	return s.routes()
}

func NewWithRealtime(service AuthService, cookieSecure bool, markets MarketAPI, lists WatchlistAPI, ingestion IngestionAPI, verifier WebhookVerification, stream http.Handler) http.Handler {
	s := &server{auth: service, cookieSecure: cookieSecure, market: markets, watchlist: lists, ingestion: ingestion, verifier: verifier, stream: stream}
	return s.routes()
}

func NewWithTradingRealtime(service AuthService, cookieSecure bool, markets MarketAPI, lists WatchlistAPI, trades TradingAPI, ingestion IngestionAPI, verifier WebhookVerification, stream http.Handler) http.Handler {
	s := &server{auth: service, cookieSecure: cookieSecure, market: markets, watchlist: lists, trading: trades, ingestion: ingestion, verifier: verifier, stream: stream}
	return s.routes()
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("POST /api/v1/auth/register", s.register)
	mux.HandleFunc("POST /api/v1/auth/login", s.login)
	mux.HandleFunc("POST /api/v1/auth/logout", s.logout)
	mux.Handle("GET /api/v1/me", s.requireAuth(http.HandlerFunc(s.me)))
	if s.market != nil {
		mux.Handle("GET /api/v1/markets/symbols", s.requireAuth(http.HandlerFunc(s.searchSymbols)))
		mux.Handle("GET /api/v1/markets/quotes/{symbol}", s.requireAuth(http.HandlerFunc(s.getQuote)))
		mux.Handle("GET /api/v1/markets/candles/{symbol}", s.requireAuth(http.HandlerFunc(s.getDailyCandles)))
	}
	if s.watchlist != nil {
		mux.Handle("GET /api/v1/watchlist", s.requireAuth(http.HandlerFunc(s.listWatchlist)))
		mux.Handle("POST /api/v1/watchlist", s.requireAuth(http.HandlerFunc(s.addWatchlist)))
		mux.Handle("PUT /api/v1/watchlist/order", s.requireAuth(http.HandlerFunc(s.reorderWatchlist)))
		mux.Handle("DELETE /api/v1/watchlist/{symbol}", s.requireAuth(http.HandlerFunc(s.deleteWatchlist)))
	}
	if s.ingestion != nil && s.verifier != nil {
		mux.HandleFunc("POST /api/v1/internal/market-data/akshare", s.ingestAkShare)
	}
	if s.stream != nil {
		mux.Handle("GET /api/v1/markets/stream", s.requireAuth(s.stream))
	}
	if s.trading != nil {
		mux.Handle("POST /api/v1/orders", s.requireAuth(http.HandlerFunc(s.createOrder)))
		mux.Handle("GET /api/v1/orders", s.requireAuth(http.HandlerFunc(s.orders)))
		mux.Handle("GET /api/v1/orders/{id}", s.requireAuth(http.HandlerFunc(s.getOrder)))
		mux.Handle("POST /api/v1/orders/{id}/cancel", s.requireAuth(http.HandlerFunc(s.cancelOrder)))
		mux.Handle("GET /api/v1/account", s.requireAuth(http.HandlerFunc(s.account)))
		mux.Handle("GET /api/v1/positions", s.requireAuth(http.HandlerFunc(s.positions)))
	}
	return s.requestContext(mux)
}

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	checker, ok := s.auth.(interface{ Health(context.Context) error })
	if !ok || checker.Health(r.Context()) != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decode(w, r, &body) {
		s.fail(w, r, http.StatusBadRequest, "invalid_request", "Request body is invalid")
		return
	}
	user, token, expires, err := s.auth.Register(r.Context(), domain.Registration{Username: body.Username, Email: body.Email, Password: body.Password})
	if err != nil {
		s.authError(w, r, err)
		return
	}
	s.setSession(w, token, expires)
	writeJSON(w, http.StatusCreated, map[string]any{"user": user})
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Identity string `json:"identity"`
		Password string `json:"password"`
	}
	if !decode(w, r, &body) {
		s.fail(w, r, http.StatusBadRequest, "invalid_request", "Request body is invalid")
		return
	}
	user, token, expires, err := s.auth.Login(r.Context(), domain.Credentials{Identity: body.Identity, Password: body.Password})
	if err != nil {
		s.authError(w, r, err)
		return
	}
	s.setSession(w, token, expires)
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		if err := s.auth.Logout(r.Context(), cookie.Value); err != nil {
			s.fail(w, r, http.StatusInternalServerError, "internal_error", "Request could not be completed")
			return
		}
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", HttpOnly: true, Secure: s.cookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: -1})
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) me(w http.ResponseWriter, r *http.Request) {
	actor := r.Context().Value(actorKey{}).(domain.Actor)
	writeJSON(w, http.StatusOK, map[string]any{"user": domain.User{ID: actor.UserID, Username: actor.Username, Email: actor.Email}})
}

func (s *server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err != nil {
			s.fail(w, r, http.StatusUnauthorized, "unauthenticated", "Please sign in")
			return
		}
		actor, err := s.auth.Authenticate(r.Context(), cookie.Value)
		if err != nil {
			s.fail(w, r, http.StatusUnauthorized, "unauthenticated", "Please sign in")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), actorKey{}, actor)))
	})
}

func (s *server) setSession(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: token, Path: "/", Expires: expires, HttpOnly: true, Secure: s.cookieSecure, SameSite: http.SameSiteLaxMode})
}

func (s *server) requestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if !requestIDPattern.MatchString(id) {
			id = newRequestID()
		}
		w.Header().Set("X-Request-ID", id)
		wrapped := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(wrapped, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
		log.Printf("request_id=%s method=%s path=%s status=%d", id, r.Method, r.URL.Path, wrapped.status)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *responseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}

func (s *server) authError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, auth.ErrInvalidInput):
		s.fail(w, r, http.StatusBadRequest, "invalid_input", "Check the submitted details")
	case errors.Is(err, auth.ErrCredentials):
		s.fail(w, r, http.StatusUnauthorized, "invalid_credentials", "Username or password is incorrect")
	case errors.Is(err, auth.ErrConflict):
		s.fail(w, r, http.StatusConflict, "identity_unavailable", "Username or email is unavailable")
	default:
		s.fail(w, r, http.StatusInternalServerError, "internal_error", "Request could not be completed")
	}
}

func (s *server) fail(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	id, _ := r.Context().Value(requestIDKey{}).(string)
	writeJSON(w, status, apiError{Code: code, Message: message, RequestID: id})
}

func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return false
	}
	return decoder.Decode(new(any)) == io.EOF
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func newRequestID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "request-unavailable"
	}
	return hex.EncodeToString(value)
}
