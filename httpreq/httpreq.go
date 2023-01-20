package httpreq

/**
 * Log HTTP-requests.
 */
import (
	"github.com/itshosted/mcore/log"
	"github.com/itshosted/webutils/middleware"
	"net/http"
)

// Write HTTP request to log.
func Use() middleware.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) bool {
		form := ""
		if r.PostForm != nil {
			form = r.PostForm.Encode()
		}
		log.Println("httpreq: %s %s %s (IP=%s)(Form=%s)", r.Proto, r.Method, r.URL, r.RemoteAddr, form)
		return true
	}
}
