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
	"webutils/report"
	"strconv"
	"time"
	"webutils/encrypt"
)

/**
 * Cookie data structure.
 *
 * This structure stores IP+UserAgent
 * so we can check if it isn't stolen by
 * another person.
 */
type Session struct {
	Id       int64
	Ldap     string
	Random   string
	Reseller string
	Ip       string
	Ua       string /* User-agent */
}

func (s *Session) Encrypt(iv string) (string, error) {
	return encrypt.EncryptBase64("aes", iv, s)
}

func (s *Session) Decrypt(in string, iv string) error {
	return encrypt.DecryptBase64("aes", iv, in, s)
}

func Expire(w http.ResponseWriter, cookie *http.Cookie) {
	cookie.Value = ""
	cookie.Expires = time.Date(1970, 1, 1, 1, 0, 0, 0, time.UTC)
	cookie.Path = "/"
	http.SetCookie(w, cookie)
}

func Read(w http.ResponseWriter, r *http.Request, proxy bool, IV string) (*Session, error) {
	sess := new(Session)
	c, e := r.Cookie("sess")
	if e != nil {
		return nil, e
	}
	e = sess.Decrypt(c.Value, IV)
	if e != nil {
		return nil, e
	}

	// Session theft protection
	if proxy {
		if sess.Ip != r.Header.Get("X-Real-IP") {
			report.Msg("[IP CHANGED] Possible stolen cookie for loginId=" + strconv.FormatInt(sess.Id, 10))
			Expire(w, c)
			return nil, errors.New("IP changed")
		}
	} else {
		if sess.Ip != r.RemoteAddr {
			report.Msg("[IP CHANGED] Possible stolen cookie for loginId=" + strconv.FormatInt(sess.Id, 10))
			Expire(w, c)
			return nil, errors.New("IP changed")
		}
	}

	if sess.Ua != r.Header.Get("User-Agent") {
		report.Msg("[UA CHANGED] Possible stolen cookie for loginId=" + strconv.FormatInt(sess.Id, 10))
		Expire(w, c)
		return nil, errors.New("UserAgent changed")
	}
	return sess, nil
}
