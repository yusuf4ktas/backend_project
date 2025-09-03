package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"strconv"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/time/rate"
)

// Creating a custom type context key to avoid collisions.
type contextKey string

const UserIDContextKey = contextKey("userID")
const RequestIDContextKey = contextKey("requestID")

func (s *Server) RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), RequestIDContextKey, requestID)

		s.logger.Info("incoming request",
			"id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)

		start := time.Now()
		rw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(rw, r.WithContext(ctx))

		s.logger.Info("finished request",
			"id", requestID,
			"status", rw.Status(),
			"bytes", rw.BytesWritten(),
			"duration", time.Since(start).String(),
		)
	})
}

func RateLimiter(next http.Handler) http.Handler {
	clients := make(map[string]*rate.Limiter)
	var mu sync.Mutex // Use a mutex to safely access the map concurrently.

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the IP address for the current user.
		ip := r.RemoteAddr

		// --- Optional improvement ---
		// For the case when needed only the host (IP) part without the port:
		//
		// host, _, err := net.SplitHostPort(r.RemoteAddr)
		// if err == nil {
		//     ip = host
		// }
		// -------------------------------------------------------

		mu.Lock()
		// Checking if there is a limiter for the IP address yet.
		if _, found := clients[ip]; !found {
			// Creating a new limiter that allows 1 request per second, with a burst of 7
			clients[ip] = rate.NewLimiter(1, 7)
		}

		if !clients[ip].Allow() {
			mu.Unlock()
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		// Extracting the token from the "Bearer <token>" format
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			http.Error(w, "Authorization header format must be Bearer {token}", http.StatusUnauthorized)
			return
		}
		tokenString := headerParts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Extract user ID from claims and add it to the request context.
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userIDFloat, ok := claims["sub"].(float64); ok {
				userID := int64(userIDFloat)
				// Creating new context with the user ID
				ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
	})
}

// AdminOnlyMiddleware is a guard that only allows users with the 'admin' role to proceed.
func (s *Server) AdminOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Get the UserID from the context that was set by the AuthMiddleware.
		userID, ok := r.Context().Value(UserIDContextKey).(int64)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		user, err := s.userService.GetByID(r.Context(), userID)
		if err != nil {
			http.Error(w, "Failed to retrieve user information", http.StatusInternalServerError)
			return
		}

		if user.Role != "admin" {
			http.Error(w, "Forbidden: This action requires admin privileges", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"path", "method", "code"}) // Labels to slice the data by.

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"path", "method"})
)

func (s *Server) PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		statusCode := rw.Status()

		httpRequestDuration.WithLabelValues(r.URL.Path, r.Method).Observe(duration)
		httpRequestsTotal.WithLabelValues(r.URL.Path, r.Method, strconv.Itoa(statusCode)).Inc()
	})
}
