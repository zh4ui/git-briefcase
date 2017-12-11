package main

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"time"
)

type GitDocityServer struct {
	gd   *GitDocity
	tmpl *template.Template
}

func NewGitDocityServer() *GitDocityServer {
	s := &GitDocityServer{}
	s.gd = NewGitDocity()
	s.tmpl = template.New("git-docity")

	http.HandleFunc("/docpack/", s.docpackHandler)
	http.HandleFunc("/", s.rootHandler)
	// TODO: serve static file here

	return s
}

func (s *GitDocityServer) Run(servingAddr string, templateDir string) {

	indexPage := filepath.Join(templateDir, "index.html")
	s.tmpl = template.Must(s.tmpl.ParseFiles(indexPage))

	err := http.ListenAndServe(servingAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func (s *GitDocityServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.errorHandler(w, r, http.StatusNotFound)
		return
	}
	s.tmpl.ExecuteTemplate(w, "index.html", s.gd)
}

func (s *GitDocityServer) docpackHandler(w http.ResponseWriter, r *http.Request) {
	log.Print(r.Method, " ", r.URL)

	pattern := regexp.MustCompile(`/docpack/([^/]+)(/.*)?`)
	matches := pattern.FindStringSubmatch(r.URL.Path)
	if matches == nil {
		// the only nil case is "/docpack/"
		// redirect to "/"
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	docname, subpath := matches[1], matches[2]
	docpack, present := s.gd.Docs[docname]
	if !present {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `docpack "%s" not found`, html.EscapeString(docname))
		return
	}

	if subpath == "" || subpath == "/" {
		subpath = docpack.IndexPage
	} else {
		// exclude leading slash
		subpath = subpath[1:]
	}

	gitdir := filepath.Join(GitDocityReposDir, docname+".git")
	gitobj, found := GitGetHashByPath(gitdir, "HEAD", subpath)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `git object "%s" not found`, html.EscapeString(subpath))
		return
	}

	content, ok := GitGetBlobContent(gitdir, gitobj.Hash)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `hash content "%s" not found`, html.EscapeString(gitobj.Hash))
		return
	}

	reader := bytes.NewReader(content)
	http.ServeContent(w, r, subpath, time.Now(), reader)
}

func (s *GitDocityServer) errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprint(w, "custom 404")
	}
}
