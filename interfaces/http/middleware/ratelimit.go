package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu       sync.Mutex
	limiters = make(map[string]*ipLimiter)
)

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	l, ok := limiters[ip]
	if !ok {
		l = &ipLimiter{limiter: rate.NewLimiter(rate.Every(time.Minute/20), 20)}
		limiters[ip] = l
	}
	l.lastSeen = time.Now()
	return l.limiter
}

// StartLimiterCleanup removes stale IP entries every 5 minutes. Call in a goroutine from main.
func StartLimiterCleanup() {
	for range time.Tick(5 * time.Minute) {
		mu.Lock()
		for ip, l := range limiters {
			if time.Since(l.lastSeen) > 10*time.Minute {
				delete(limiters, ip)
			}
		}
		mu.Unlock()
	}
}

// RateLimitMiddleware applies a token-bucket rate limit of 20 req/min per IP.
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := RealIP(c)
		if !getLimiter(ip).Allow() {
			zap.L().Warn("rate limit hit", zap.String("ip", ip))
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Demasiadas solicitudes. Intentá más tarde."})
			return
		}
		c.Next()
	}
}
