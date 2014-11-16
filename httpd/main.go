package httpd
/**
 * Lazy utility methods for HTTP-server.
 */
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"webutils/report"
)

type DefaultResponse struct {
	Status bool   `json:"status"`
	Text   string `json:"text"`
}

func Reply(status bool, text string) DefaultResponse {
	return DefaultResponse{status, text}
}

// Write v as string to w
func FlushJson(w http.ResponseWriter, v interface{}) error {
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	fmt.Fprint(w, string(b))
	return nil
}

// Read and unmarshal request input
func ReadInput(r *http.Request, out interface{}) error {
	defer r.Body.Close()

	body, e := ioutil.ReadAll(r.Body)
	if e != nil {
		return e
	}

	if e := json.Unmarshal(body, out); e != nil {
		return e
	}
	return nil
}

// Read and unmarshal response input
func ReadOutput(r *http.Response, out interface{}) error {
	defer r.Body.Close()

	body, e := ioutil.ReadAll(r.Body)
	if e != nil {
		return e
	}

	if e := json.Unmarshal(body, out); e != nil {
		return e
	}
	return nil
}

// Write msg as error and report e to log
func Error(w http.ResponseWriter, e error, msg string) {
	report.Err(e)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)
	if e := FlushJson(w, Reply(false, msg)); e != nil {
		panic(e)
	}
}
