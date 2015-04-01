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

var Cache *lru.Cache
var Burst float64

func init() {
	/* LRU cache for a max of 1000 entries */
	Cache = lru.New(1000)
}

// return true on ratelimit reached
func isRequestOk(Addr string, Burst float64) bool {
	ip := strings.Split(Addr, ":")[0]

	item, newEntry := Cache.Get(ip)
	if !newEntry {
		fmt.Println("Entry not found in cache, adding")

		item = bucket.New(1.0, Burst)
		Cache.Add(ip, item)
		return false
	}

	/* Cast cache item */
	c := item.(*bucket.Bucket)
	ok, _ := c.Request(1.0)
	return ok
}

func Use(burst float64) middleware.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) bool {
		ip := r.RemoteAddr

		ok := isRequestOk(ip, burst)
		if !ok {
			w.WriteHeader(429)
			if e := httpd.FlushJson(w, httpd.Reply(false, "Ratelimit reached")); e != nil {
				httpd.Error(w, e, "Flush failed")
			}
			return false
		}
		return true
	}
}
