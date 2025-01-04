package httputils

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
	"gurch101.github.io/go-web/pkg/parser"
)

var ErrPanic = errors.New("panic")

// LoggingMiddleware logs the request and response details.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		RequestID := r.Header.Get("X-Request-ID")
		ctx := r.Context()

		if RequestID != "" {
			ctx = context.WithValue(ctx, LogRequestIDKey, "ext-"+RequestID)
		} else {
			id := uuid.New()
			ctx = context.WithValue(ctx, LogRequestIDKey, id.String())
		}

		slog.InfoContext(ctx, "request started")
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

		duration := time.Since(start)

		slog.InfoContext(
			ctx,
			"request completed",
			"request_method", r.Method,
			"request_url", r.URL.String(),
			"duration", duration.Milliseconds(),
		)
	})
}

// RecoveryMiddleware recovers from panics and sends a 500 Internal Server Error response.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				ServerErrorResponse(w, r, fmt.Errorf("%w: %s", ErrPanic, err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type RateLimitConfig struct {
	enabled bool
	rate    float64
	burst   int
}

const (
	defaultRateLimitRate = 10

	defaultRateLimitBurst = 20
)

func getRateLimitConfig() *RateLimitConfig {
	rateLimitConfig := &RateLimitConfig{
		enabled: parser.ParseEnvBool("RATE_LIMIT_ENABLED", true),
		rate:    defaultRateLimitRate,
		burst:   defaultRateLimitBurst,
	}
	if !rateLimitConfig.enabled {
		return rateLimitConfig
	}

	rateLimit, err := parser.ParseEnvFloat64("RATE_LIMIT_RATE", rateLimitConfig.rate)
	if err != nil {
		panic(err)
	}

	rateLimitConfig.rate = rateLimit

	burst, err := parser.ParseEnvInt("RATE_LIMIT_BURST", rateLimitConfig.burst)
	if err != nil {
		panic(err)
	}

	rateLimitConfig.burst = burst

	return rateLimitConfig
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	rateLimitConfig := getRateLimitConfig()

	if !rateLimitConfig.enabled {
		return next
	}

	slog.Info("rate limit middleware enabled", "rate", rateLimitConfig.rate, "burst", rateLimitConfig.burst)

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ServerErrorResponse(w, r, fmt.Errorf("could not parse remote address: %w", err))
		}

		mu.Lock()
		if _, ok := clients[ip]; !ok {
			limiter := rate.NewLimiter(
				rate.Limit(rateLimitConfig.rate),
				rateLimitConfig.burst,
			)
			clients[ip] = &client{limiter: limiter, lastSeen: time.Now()}
		} else {
			clients[ip].lastSeen = time.Now()
		}

		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			RateLimitExceededResponse(w, r)

			return
		}

		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

type UnauthorizedRedirector func(w http.ResponseWriter, r *http.Request, destURL string)

func GetStateAwareAuthenticationMiddleware(_ UnauthorizedRedirector) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}
