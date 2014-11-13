package utils

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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"webutils/report"
	"strconv"
	"time"
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
	if len(iv) != 32 {
		return "", errors.New("IV must be random 32-chars")
	}
	key := []byte(iv)
	json, e := json.Marshal(s)
	if e != nil {
		return "", e
	}

	raw, e := s.encrypt(key, []byte(json))
	if e != nil {
		return "", e
	}
	return url.QueryEscape(base64.StdEncoding.EncodeToString(raw)), nil
}

func (s *Session) Decrypt(in string, iv string) error {
	if len(iv) != 32 {
		return "", errors.New("IV must be random 32-chars")
	}
	key := []byte(iv)

	b, e := url.QueryUnescape(in)
	if e != nil {
		return e
	}
	str, e := base64.StdEncoding.DecodeString(b)
	if e != nil {
		return e
	}
	raw, e := s.decrypt(key, str)
	if e != nil {
		return e
	}

	if e := json.Unmarshal(raw, s); e != nil {
		return e
	}
	return nil
}

func (s *Session) encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

func (s *Session) decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ExpireSession(w http.ResponseWriter, cookie *http.Cookie) {
	cookie.Value = ""
	cookie.Expires = time.Date(1970, 1, 1, 1, 0, 0, 0, time.UTC)
	cookie.Path = "/"
	http.SetCookie(w, cookie)
}

func ReadSession(w http.ResponseWriter, r *http.Request, proxy bool) (*Session, error) {
	sess := new(Session)
	c, e := r.Cookie("sess")
	if e != nil {
		return nil, e
	}
	e = sess.Decrypt(c.Value)
	if e != nil {
		return nil, e
	}

	// Session theft protection
	if proxy {
		if sess.Ip != r.Header.Get("X-Real-IP") {
			report.Msg("[IP CHANGED] Possible stolen cookie for loginId=" + strconv.FormatInt(sess.Id, 10))
			ExpireSession(w, c)
			return nil, errors.New("IP changed")
		}
	} else {
		if sess.Ip != r.RemoteAddr {
			report.Msg("[IP CHANGED] Possible stolen cookie for loginId=" + strconv.FormatInt(sess.Id, 10))
			ExpireSession(w, c)
			return nil, errors.New("IP changed")
		}
	}

	if sess.Ua != r.Header.Get("User-Agent") {
		report.Msg("[UA CHANGED] Possible stolen cookie for loginId=" + strconv.FormatInt(sess.Id, 10))
		ExpireSession(w, c)
		return nil, errors.New("UserAgent changed")
	}
	return sess, nil
}
