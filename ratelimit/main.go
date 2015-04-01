package ratelimit

/**
 * HTTP Ratelimiter
 */
import (
	"github.com/golang/groupcache/lru"
	"github.com/xsnews/webutils/httpd"
	"github.com/xsnews/webutils/middleware"
	"github.com/xsnews/webutils/ratelimit/bucket"
	"net/http"
	"strings"
	"time"
)

const (
	StatusRateLimit = 429
)

var DelayThreshold = 10

var Cache *lru.Cache

func init() {
	/* LRU cache for a max of 1000 entries */
	Cache = lru.New(1000)
}

/* Returns http status code */
func isRequestOk(addr string, rate float64, burst float64, delay time.Duration) int {
	ip := strings.Split(addr, ":")[0]

	item, newEntry := Cache.Get(ip)
	if !newEntry {
		item = bucket.New(rate, burst, delay)
		Cache.Add(ip, item)
		return http.StatusOK
	}

	/* Cast to Bucket */
	c := item.(*bucket.Bucket)

	/* Request a token from bucket */
	ok, _ := c.Request(1.0)
	if !ok {
		/* Did we exceed our ratelimit threshold? */
		if c.DelayCounter >= DelayThreshold {
			return http.StatusServiceUnavailable
		}

		/* Ratelimit */
		return StatusRateLimit
	}

	/* Everything is OK */
	return http.StatusOK
}

func Use(fillrate float64, capacity float64, delay time.Duration) middleware.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) bool {
		code := isRequestOk(r.RemoteAddr, fillrate, capacity, delay)
		switch code {
		case StatusRateLimit:
			/* Ratelimit request */
			w.WriteHeader(StatusRateLimit)
			if e := httpd.FlushJson(w, httpd.Reply(false, "Ratelimit reached")); e != nil {
				httpd.Error(w, e, "Flush failed")
			}
			return false
		case http.StatusServiceUnavailable:
			/* Max number of ratelimits exceeded, make service unavailable for this IP */
			w.WriteHeader(http.StatusServiceUnavailable)
			if e := httpd.FlushJson(w, httpd.Reply(false, "Service temporarily unavailable")); e != nil {
				httpd.Error(w, e, "Flush failed")
			}
			return false
		default:
			return true
		}
	}
}
