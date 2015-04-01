package ratelimit

/**
 * HTTP Ratelimiter
 */
import (
	"fmt"
	"github.com/golang/groupcache/lru"
	"github.com/xsnews/webutils/httpd"
	"github.com/xsnews/webutils/middleware"
	"github.com/xsnews/webutils/ratelimit/bucket"
	"net/http"
	"strings"
)

type Limit struct {
	Burst  float64
	Proxy  bool   /* use proxy IP */
	Prefix string /* redis key prefix */
}

var Cache *lru.Cache

func init() {
	/* LRU cache for a max of 1000 entries */
	Cache = lru.New(1000)
}

// return true on ratelimit reached
func isRateLimitReached(Addr string, Prefix string, Burst float64) bool {
	ip := strings.Split(Addr, ":")[0]
	key := Prefix + "_" + ip

	item, ok := Cache.Get(key)
	if !ok {
		fmt.Println("Entry not found in cache, adding")

		item = bucket.New(1.0, Burst)
		Cache.Add(key, item)
		return false
	}

	/* Cast cache item */
	c := item.(*bucket.Bucket)
	requestOk, _ := c.Request(1.0)
	if requestOk {
		return false
	} else {
		return true
	}
}

func Use(limit Limit) middleware.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) bool {
		ip := r.RemoteAddr
		if limit.Proxy {
			ip = r.Header.Get("X-Real-IP")
		}

		doLimit := isRateLimitReached(ip, limit.Prefix, limit.Burst)
		if doLimit {
			w.WriteHeader(429)
			if e := httpd.FlushJson(w, httpd.Reply(false, "Ratelimit reached")); e != nil {
				httpd.Error(w, e, "Flush failed")
			}
			return false
		}
		return true
	}
}
