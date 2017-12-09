package main

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
)

type GitBriefcaseServer struct {
	gb   *GitBriefcase
	tmpl *template.Template
}

func NewGitBriefcaseServer() *GitBriefcaseServer {
	s := &GitBriefcaseServer{}
	s.gb = NewGitBriefcase()
	s.tmpl = template.New("git-briefcase")

	http.HandleFunc("/docpack/", s.docpackHandler)
	http.HandleFunc("/", s.rootHandler)
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

func (s *GitBriefcaseServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.errorHandler(w, r, http.StatusNotFound)
		return
	}
	s.tmpl.ExecuteTemplate(w, "index.html", s.gb)
}

func (s *GitBriefcaseServer) docpackHandler(w http.ResponseWriter, r *http.Request) {
	pattern := regexp.MustCompile(`/docpack/([^/]+)(/.*)?`)
	matches := pattern.FindStringSubmatch(r.URL.Path)
	if matches == nil {
		// the only nil case is "/docpack/"
		// redirect to "/"
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	docname, subpath := matches[1], matches[2]
	docpack, present := s.gb.Docs[docname]
	if !present {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `docpack "%s" not found`, html.EscapeString(docname))
		return
	}
	_ = subpath
	_ = docpack

	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func (s *GitBriefcaseServer) errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprint(w, "custom 404")
	}
}
