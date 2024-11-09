package main

import (
	cyoi "cyoi"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	fileName := flag.String(
		"file",
		"gopher.json",
		"Use to point to the story in .json format that will be retold",
	)
	port := flag.Int("port", 3000, "the port to start the CYOA web application on")
	flag.Parse()

	jsonFile, err := os.Open(*fileName)
	if err != nil {
		fmt.Printf("Failed to open file %q, %s", fileName, err)
	}
	defer jsonFile.Close()

	mux := http.NewServeMux()

	story, err := cyoi.JsonStory(jsonFile)
	if err != nil {
		fmt.Printf("Failed to decode file %q, %s", fileName, err)
	}
	t, _ := template.New("").Parse(tpl)
	h := cyoi.NewHandler(story, cyoi.WithTemplate(t), cyoi.WithPathFunc(PathFunc))
	mux.HandleFunc("/story/", h.ServeHTTP)
	fmt.Printf("Starting the web application on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), mux))

}

func PathFunc(r *http.Request) string {
	path := strings.TrimPrefix(r.URL.Path, "/story/")
	fmt.Println(path)
	if path == "" {
		path = "intro"
	}
	return path
}

var tpl = `
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
            <li><a href="/story/{{.Chapter}}">{{.Text}}</a></li>
            {{end}}
        </ul>
    </body>
</html>`
