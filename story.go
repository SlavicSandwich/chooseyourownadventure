package chooseyourownadventure

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

func init() {
	tpl = template.Must(template.New("").Parse(defaultHandlerTemplate))
}

var tpl *template.Template

var defaultHandlerTemplate = `
<!DOCTYPE html>
<html>
    <head>
        <meta charset="UTF-8">
        <title>Choose your own adventure</title>
    </head>
    <body>
        <h1>{{.Title}}</h1>
        {{range .Paragraphs}}
        <p>{{.}}</p>
        {{end}}
        <ul>
            {{range .Options}}
            <li><a href="/{{.Chapter}}">{{.Text}}</a></li>
            {{end}}
        </ul>
    </body>
</html>`

func JsonStory(r io.Reader) (Story, error) {
	decoder := json.NewDecoder(r)
	var story Story
	if err := decoder.Decode(&story); err != nil {
		return nil, err
	}
	return story, nil
}

type Story map[string]StoryArc

type StoryArc struct {
	Title      string   `json:"title"`
	Paragraphs []string `json:"story"`
	Options    []Option `json:"options"`
}

type Option struct {
	Text    string `json:"text"`
	Chapter string `json:"arc"`
}

type HandlerOption func(h *handler)

func WithTemplate(t *template.Template) HandlerOption {
	return func(h *handler) {
		h.template = t
	}
}

func WithPathFunc(fn func(r *http.Request) string) HandlerOption {
	return func(h *handler) {
		h.pathFunc = fn
	}
}

func NewHandler(story Story, opts ...HandlerOption) http.Handler {
	h := handler{story: story, template: tpl, pathFunc: defaultPathFunc}
	for _, opt := range opts {
		opt(&h)
	}
	return h
}

type handler struct {
	story    Story
	template *template.Template
	pathFunc func(r *http.Request) string
}

func defaultPathFunc(r *http.Request) string {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "intro"
	}
	return path
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := h.pathFunc(r)
	if chapter, ok := h.story[path]; ok {
		err := h.template.Execute(w, chapter)
		if err != nil {
			log.Printf("%v", err)
			http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "Chapter not found.", http.StatusNotFound)
}
