package ratelimit
/**
 * HTTP Ratelimiter. Limit the amount of HTTP-Requests per second.
 * What we try to solve?
 * - Warn about abusive servers (Log if a IP comes close to a ratelimit)
 * - Block requests if limit exceeded (HTTP 429)
 * - Block requests if limit is ignored DelayTreshHold times (HTTP 503)
 */
import (
	"github.com/golang/groupcache/lru"
	"github.com/xsnews/webutils/httpd"
	"github.com/xsnews/webutils/middleware"
	"github.com/xsnews/webutils/ratelimit/bucket"
	"net/http"
	"strings"
	"time"
	"fmt"
)

// HTTP StatusCode for Ratelimit
const (
	StatusRateLimit     = 429
	StatusRateLimitText = "Too Many Requests"
)

var (
	DelayThreshold int = 10   /* Max hits after ratelimit exceeded before making service unavailable */
	Delay          int = 10   /* Seconds to delay after DelayTreshold */
	CacheSize      int = 1000 /* Max IPs we ratelimit */
)

// Fixed size queue. If cache
// gets bigger than CacheSize the LRU (Least Recently Used)
// item is deleted first.
var cache *lru.Cache

// Check IP ratelimit and return HTTP-statuscode.
func check(addr string, rate float64, burst float64) int {
	ip := strings.Split(addr, ":")[0]

	request, isNewRequest := cache.Get(ip)
	if !isNewRequest {
		request = bucket.New(rate, burst, time.Duration(Delay)*time.Second)
		cache.Add(ip, request)
		return http.StatusOK
	}

	c := request.(*bucket.Bucket)
	ok := c.Request(1.0)

	if !ok {
		if c.DelayCounter >= DelayThreshold {
			// Abusive Microservice keeps flooding us.
			// Change HTTP-statuscode to get attention!
			return http.StatusServiceUnavailable
		}
		return StatusRateLimit
	}
	return http.StatusOK
}

// Limit the amount of requests one IP can do per second.
//
// fillrate = Amount of requests allowed in one second
// capacity = Amount of 'extra'(burst) requests a-top fillrate allowed
// delay = Duration before changing http status 429 to 503
//
// If the fillrate+capacity are overloaded a HTTP 429 is returned
// If the callee keeps firing requests the 429 is changed into a 503
// if delay is passed. 
func Use(fillrate float64, capacity float64) middleware.HandlerFunc {
	cache = lru.New(CacheSize)

	return func(w http.ResponseWriter, r *http.Request) bool {
		httpCode := check(r.RemoteAddr, fillrate, capacity)
		switch httpCode {
		case StatusRateLimit:
			/* Ratelimit request */
			fmt.Println("CRIT: Ratelimit HTTP-request for IP=" + r.RemoteAddr + " (request dropped)")
			w.WriteHeader(StatusRateLimit)
			if e := httpd.FlushJson(w, httpd.Reply(false, StatusRateLimitText)); e != nil {
				httpd.Error(w, e, "Flush failed")
			}
			return false
		case http.StatusServiceUnavailable:
			/* Max number of ratelimits exceeded, make service unavailable for this IP */
			fmt.Println("CRIT: Ratelimit ignored by IP=" + r.RemoteAddr + " (request dropped)")
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
