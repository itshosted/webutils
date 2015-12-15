package session

/**
 * Client-side session cookie.
 *  By offering cookie wit AES256 encoded content
 *
 * Why would you want this?
 *  To have stateless server code so you can
 *  send an user to multiple HTTP-servers that
 *  don't need talk to each other.
 */
import (
	"errors"
	"net/http"
	"time"
	"strings"
	"github.com/itshosted/webutils/encrypt"
	"github.com/itshosted/mcore/log"
	"github.com/itshosted/webutils/str"
)

const COOKIE = "sess"

/**
 * Cookie data structure.
 *
 * This structure stores IP+UserAgent
 * so we can check if it isn't stolen by
 * another person.
 */
type Session struct {
	Random   string      /* Jitter so cookie always changes */
	Ip       string      /* Visitor IP */
	Ua       string      /* Visitor User-agent */
	More     interface{} /* More data */

	expires   time.Time   /* Expiration time */
	iv        string
	httpsOnly bool
}

// Expire cookie
func Expire(w http.ResponseWriter) {
	cookie := new(http.Cookie)
	cookie.Name = COOKIE
	cookie.Value = ""
	cookie.Expires = time.Date(1970, 1, 1, 1, 0, 0, 0, time.UTC)
	cookie.Path = "/"
	http.SetCookie(w, cookie)
}

// Get visitor IP
// Force through X-Real-IP as Go microservices
// should be called through reverse proxy
func ip(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		log.Println(
			"WARN: Reverse-proxy does not supply X-Real-IP field! (https://github.com/xsnews/webutils/tree/master/session)",
		)
		// Fallback
		ip = r.RemoteAddr
	}

	return ip
}

// Read request and return session
// Take writer to expire cookie on theft
func Get(w http.ResponseWriter, r *http.Request, iv string, more interface{}) error {
	sess := new(Session)
	sess.More = more

	c, e := r.Cookie(COOKIE)
	if e != nil {
		if e == http.ErrNoCookie {
			return nil
		}
		return e
	}
	if e := encrypt.DecryptBase64("aes", iv, c.Value, sess); e != nil {
		return e
	}

	// Session theft protection
	ip := ip(r)
	if sess.Ip != ip {
		log.Println("WARN: IP changed, cookie expired for IP=%s/%s", sess.Ip, ip)
		Expire(w)
		return errors.New("IP changed")
	}
	if sess.Ua != r.Header.Get("User-Agent") {
		log.Println(
			"WARN: User-Agent changed, cookie expired for IP=%s/%s (UA=%s/%s)",
			sess.Ip, ip, sess.Ua, r.Header.Get("User-Agent"),
		)
		Expire(w)
		return errors.New("UserAgent changed")
	}

	return nil
}

// Security notice:
// - By default HttpOnly is set, so JS can't see the cookie
// - If httpsOnly is set the browser only reads the cookie through https
func New(w http.ResponseWriter, r *http.Request, expires time.Time, iv string, httpsOnly bool, more interface{}) error {
	s := Session{
		Random: str.RandText(10),
		Ip: ip(r),
		Ua: r.Header.Get("User-Agent"),
		More: more,
	}

	c, e := encrypt.EncryptBase64("aes", iv, s)
	if e != nil {
		return e
	}

	cookie := new(http.Cookie)
	cookie.Name = COOKIE
	cookie.Value = c
	cookie.Path = "/"
	cookie.Domain = strings.Split(r.Header.Get("Host"), ":")[0]
	cookie.Expires = s.expires
	if httpsOnly {
		cookie.Secure = true
	}
	cookie.HttpOnly = true
	http.SetCookie(w, cookie)
	return nil
}
