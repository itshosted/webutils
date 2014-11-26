package encrypt

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
)

func EncryptBase64(enc string, iv string, in interface{}) (string, error) {
	var (
		raw []byte
		e   error
	)
	if len(iv) != 32 {
		return "", errors.New("IV must be random 32-chars")
	}
	key := []byte(iv)
	raw, e = json.Marshal(in)
	if e != nil {
		return "", e
	}

	if enc == "aes" {
		raw, e = aesEncrypt(key, []byte(raw))
	} else {
		return "", errors.New("Unsupported encoding type=" + enc)
	}
	if e != nil {
		return "", e
	}
	return url.QueryEscape(base64.StdEncoding.EncodeToString(raw)), nil
}

func DecryptBase64(enc string, iv string, in string, out interface{}) error {
	var (
		e   error
		str string
		b   []byte
	)
	if len(iv) != 32 {
		return errors.New("IV must be random 32-chars")
	}
	key := []byte(iv)

	str, e = url.QueryUnescape(in)
	if e != nil {
		return e
	}
	b, e = base64.StdEncoding.DecodeString(str)
	if e != nil {
		return e
	}

	if enc == "aes" {
		b, e = aesDecrypt(key, b)
	} else {
		return errors.New("Unsupported encoding type=" + enc)
	}
	if e != nil {
		return e
	}

	if e := json.Unmarshal(b, out); e != nil {
		return e
	}
	return nil
}
