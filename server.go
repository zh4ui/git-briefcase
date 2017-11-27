package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type GitBriefcaseServer struct {
	gb   *GitBriefcase
	tmpl *template.Template
}

func NewGitBriefcaseServer() *GitBriefcaseServer {
	s := &GitBriefcaseServer{}
	s.gb = NewGitBriefcase()
	s.tmpl = template.New("git-briefcase")

	http.Handle("/", s)
	// TODO: serve static file here

	return s
}

func (s *GitBriefcaseServer) Run(servingAddr string, templateDir string) {

	indexPage := filepath.Join(templateDir, "index.html")
	s.tmpl = template.Must(s.tmpl.ParseFiles(indexPage))

	err := http.ListenAndServe(servingAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func (s *GitBriefcaseServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	s.tmpl.ExecuteTemplate(w, "index.html", s.gb)
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprint(w, "custom 404")
	}
}
