package safehttp

/**
 * Require admin privilege unless excluded.
 *
 * Simple solution to prevent 'forgetting'
 *  permissions to become a problem.
 */
import (
	"net/http"
	"github.com/xsnews/webutils/httpd"
	"github.com/xsnews/webutils/middleware"
	"github.com/xsnews/webutils/report"
	"github.com/xsnews/webutils/session"
)

var (
	rules map[string]Rule
)

type Rule struct {
	Session bool
}

func Add(path string, rule Rule) {
	if rules == nil {
		rules = make(map[string]Rule)
	}
	rules[path] = rule
}

// Check session for -admin -loggedin
func check(s *session.Session) (bool, bool) {
	if s == nil {
		return false, false
	}
	if s.Reseller == "xsnews" {
		return true, true
	}
	if s.Id > 0 {
		return false, true
	}
	return false, false
}

func Use(proxy bool, IV string) middleware.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) bool {
		rule, hasRule := rules[r.URL.Path]
		if hasRule && rule.Session == false {
			return true
		}

		s, e := session.Read(w, r, proxy, IV)
		if e != nil {
			// Ignore session decode error
			// on error user is considered not logged in
			report.Debug(e.Error())
		}

		isAdmin, isLoggedIn := check(s)
		if hasRule {
			if rule.Session && isLoggedIn {
				return true
			}
		}

		if isAdmin {
			// Admin is always allowed
			return true
		}
		report.Msg("Blocking HTTP-request URL=" + r.URL.Path + " for IP=" + r.RemoteAddr)
		w.WriteHeader(401)
		if e := httpd.FlushJson(w, httpd.Reply(false, "Insufficient privileges")); e != nil {
			httpd.Error(w, e, "Flush failed")
		}
		return false
	}
}
