package muxdoc

/**
 * Simple MUX-wrapper to easily create
 * documentation for the API.
 */
import (
	"bytes"
	"net/http"

	"gopkg.in/yaml.v2"
)

type MuxDoc struct {
	Title string
	Desc  string
	Meta  string

	Mux  *http.ServeMux
	urls map[string]string
}

// Add URL to mux+docu
func (m *MuxDoc) Add(url string, fn func(http.ResponseWriter, *http.Request), comment string) {
	if m.Mux == nil {
		m.Mux = http.NewServeMux()
		m.urls = make(map[string]string)
	}
	m.urls[url] = comment
	m.Mux.HandleFunc(url, fn)
}

// Create documentation
func (m *MuxDoc) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("<html><head><title>" + m.Title + "</title><link href=\"//maxcdn.bootstrapcdn.com/bootstrap/3.2.0/css/bootstrap.min.css\" rel=\"stylesheet\">")
	buffer.WriteString("</head><body><div class=\"container\">")
	buffer.WriteString("<div class=\"page-header\"><h1>" + m.Title + "</h1>")
	buffer.WriteString("<p>" + m.Desc + "</p>")
	buffer.WriteString(m.Meta)
	buffer.WriteString("</div><h2>Routes</h2><table class=\"table table-striped\"><thead><tr><th>URL</th><th>Comment</th></tr></thead>")
	for url, comment := range m.urls {
		buffer.WriteString("<tr><td><a href=\"" + url + "\">" + url + "</td><td>" + comment + "</td></tr>")
	}
	buffer.WriteString("</table></div></body></html>")

	return buffer.String()
}

// ToYAML returns the same data as your original String() but in YAML form.
func (m *MuxDoc) ToYAML() (string, error) {
	// helper types for marshalling
	type urlItem struct {
		URL     string `yaml:"url"`
		Comment string `yaml:"comment"`
	}
	out := struct {
		Title       string    `yaml:"title"`
		Description string    `yaml:"description"`
		Meta        string    `yaml:"meta"`
		URL         []urlItem `yaml:"url"`
	}{
		Title:       m.Title,
		Description: m.Desc,
		Meta:        m.Meta,
		URL:         make([]urlItem, 0, len(m.urls)),
	}

	// collect URL/comment pairs
	for u, c := range m.urls {
		out.URL = append(out.URL, urlItem{URL: u, Comment: c})
	}

	// marshal to YAML
	buf := &bytes.Buffer{}
	enc := yaml.NewEncoder(buf)

	if err := enc.Encode(out); err != nil {
		return "", err
	}
	return buf.String(), nil
}
