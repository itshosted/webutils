package validate

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type cmd struct {
	Values interface{} `json:"values"`
	Rules  interface{} `json:"rules"`
}

var (
	validateUrl string
)

func Init(url string) {
	validateUrl = url
}

func Sanitize(r *http.Request, rules interface{}, output interface{}) (string, error) {
	var (
		input []byte
		e     error
	)

	// Input
	defer r.Body.Close()
	input, e = ioutil.ReadAll(r.Body)
	if e != nil {
		return "", e
	}

	// Forward
	c := new(cmd)
	c.Values = input
	c.Rules = rules

	j, e := json.Marshal(c)
	if e != nil {
		return "", e
	}

	res, e := http.Post(
		validateUrl, "application/json",
		bytes.NewBuffer(j),
	)
	if e != nil {
		return "", e
	}
	defer res.Body.Close()
	sane, e := ioutil.ReadAll(res.Body)
	if e != nil {
		return "", e
	}
	if res.StatusCode != 200 {
		return string(sane), errors.New("Received non-200 HTTP-status from validated")
	}

	// Output
	if e := json.Unmarshal(sane, &output); e != nil {
		return "", e
	}
	return "", nil
}
