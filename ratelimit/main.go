package ratelimit
/**
 * HTTP Ratelimiter
 * Use Redis to limit amount of
 *  HTTP-conns so service isn't abused
 */
import (
	"github.com/garyburd/redigo/redis"
	"net/http"
	"webutils/report"
	"webutils/middleware"
	"webutils/httpd"
	"strconv"
	"strings"
)

type Limit struct {
	RatelimitTime int /* secs */
	RatelimitMax  int /* max reqs */
	Proxy bool /* use proxy IP */
	Prefix string /* redis key prefix */
}

var (
	_pool *redis.Pool
)

func SetRedis(pool *redis.Pool) {
	_pool = pool
}

func check(Ip string, Prefix string, Expire int, Max int) (bool, error) {
	var (
		e     error
		count int
	)
	if strings.Index(Ip, ":") != -1 {
		Ip = Ip[:strings.Index(Ip, ":")]
	}
	if (_pool == nil) {
		panic("DevErr: Forgot to call SetRedis(pool)")
	}
	conn := _pool.Get()
	defer conn.Close()
	key := "RATELIMIT_" + Prefix + "_" + Ip

	count, e = redis.Int(conn.Do("INCR", key))
	if e != nil {
		return false, e
	}
	if count > Max {
		report.Msg("Ratelimit reached for IP=" + Ip + " with prefix=" + Prefix)
		return true, nil
	}

	if _, e := conn.Do("EXPIRE", key, Expire); e != nil {
		return false, e
	}
	report.Debug("Increase ratelimit for IP=" + Ip + " to=" + strconv.FormatInt(int64(count), 10))
	return false, nil
}

func Use(limit Limit) middleware.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) bool {
		ip := r.RemoteAddr
		if (limit.Proxy) {
			ip = r.Header.Get("X-Real-IP")
		}
		max, e := check(ip, limit.Prefix, limit.RatelimitTime, limit.RatelimitMax)
		if e != nil {
			// Report error and continue
			// (accepting so Redis down doesn't mean service 100% down)
			report.Err(e)
		}
		if max {
			w.WriteHeader(429)
			if e := httpd.FlushJson(w, httpd.Reply(false, "Ratelimit reached")); e != nil {
				httpd.Error(w, e, "Flush failed")
			}
			return false
		}
		return true
	}
}
