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

/* The ratelimit HTTP status code is not available in golang's HTTP library */
const (
	StatusRateLimit     = 429
	StatusRateLimitText = "Too Many Requests"
)

type RatelimitConfig struct {
	Delay          time.Duration /* Delay after ratelimit exceeded */
	DelayThreshold int           /* Max hits after ratelimit exceeded before making service unavailable */
	CacheSize      int           /* Max connections we ratelimit based on LRU cache */
}

var Config *RatelimitConfig
var Cache *lru.Cache

func init() {
	/* Set default (sane) ratelimit values */
	Config = &RatelimitConfig{
		DelayThreshold: 10,
		CacheSize:      1000,
		Delay:          time.Second * 3,
	}
}

/* Returns http status code */
func isRequestOk(addr string, rate float64, burst float64) int {
	ip := strings.Split(addr, ":")[0]

	request, isNewRequest := Cache.Get(ip)
	if !isNewRequest {
		request = bucket.New(rate, burst, Config.Delay)
		Cache.Add(ip, request)
		return http.StatusOK
	}

	/* Cast to Bucket */
	c := request.(*bucket.Bucket)

	/* Request a token from bucket */
	ok, _ := c.Request(1.0)
	if !ok {
		/* Did we exceed our ratelimit threshold? */
		if c.DelayCounter >= Config.DelayThreshold {
			return http.StatusServiceUnavailable
		}

		/* Ratelimit */
		return StatusRateLimit
	}

	/* Everything is OK */
	return http.StatusOK
}

func Use(fillrate float64, capacity float64) middleware.HandlerFunc {
	/* Initialise LRU cache */
	Cache = lru.New(Config.CacheSize)

	return func(w http.ResponseWriter, r *http.Request) bool {
		httpCode := isRequestOk(r.RemoteAddr, fillrate, capacity)
		switch httpCode {
		case StatusRateLimit:
			/* Ratelimit request */
			w.WriteHeader(StatusRateLimit)
			if e := httpd.FlushJson(w, httpd.Reply(false, StatusRateLimitText)); e != nil {
				httpd.Error(w, e, "Flush failed")
			}
			return false
		case http.StatusServiceUnavailable:
			/* Max number of ratelimits exceeded, make service unavailable for this IP */
			w.WriteHeader(http.StatusServiceUnavailable)
			if e := httpd.FlushJson(w, httpd.Reply(false, http.StatusText(http.StatusServiceUnavailable))); e != nil {
				httpd.Error(w, e, "Flush failed")
			}
			return false
		default:
			return true
		}
	}
}
