package main

import (
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

func (s *GitBriefcaseServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.tmpl.ExecuteTemplate(w, "index.html", s.gb)
}
