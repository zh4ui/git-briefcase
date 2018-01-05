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

	cache "github.com/patrickmn/go-cache"
)

type DocityServer struct {
	home       *DocityHome
	tmpl       *template.Template
	staticDir  string
	hotGitObjs *cache.Cache
}

func NewDocityServer(staticDir string) *DocityServer {
	s := &DocityServer{}
	s.staticDir = staticDir
	s.home = NewDocityHome()
	s.tmpl = template.New("git-docity")

	// 5 minutes expiration and 60 minutes purge period
	s.hotGitObjs = cache.New(10*time.Minute, 60*time.Minute)

	assetsDir := filepath.Join(s.staticDir, "assets")
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))))

	http.HandleFunc("/view/", s.viewHandler)
	http.HandleFunc("/repo/", s.repoHandler)
	http.HandleFunc("/", s.rootHandler)

	return s
}

func (s *DocityServer) Run(servingAddr string) {
	templatesDir := filepath.Join(s.staticDir, "templates")
	indexPage := filepath.Join(templatesDir, "index.gohtml")
	s.tmpl = template.Must(s.tmpl.ParseFiles(indexPage))

	err := http.ListenAndServe(servingAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func (s *DocityServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// XXX re-compile tempalte for each request,
	// should turn off for release version
	s.tmpl = template.New("git-docity")
	templatesDir := filepath.Join(s.staticDir, "templates")
	indexPage := filepath.Join(templatesDir, "index.gohtml")
	s.tmpl = template.Must(s.tmpl.ParseFiles(indexPage))
	s.tmpl.ExecuteTemplate(w, "index.gohtml", s.home)
}

func (s *DocityServer) viewHandler(w http.ResponseWriter, r *http.Request) {
	log.Print(r.Method, " ", r.URL)

	if r.URL.Path == "/view/" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	/*
		pattern := regexp.MustCompile(`/view/([^/]+)(/.*)?`)
		if matches == nil {
			// the only nil case is "/view/"
			// redirect to "/"
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		docname, subpath := matches[1], matches[2]
		docpack, present := s.home.Docs[docname]
	*/

	pattern := regexp.MustCompile(`/view/([^/]+)(/.*)?`)
	matches := pattern.FindStringSubmatch(r.URL.Path)
	if matches == nil {
		panic("impossible")
	}
	docrepo, subpath := matches[1], matches[2]
	docpack, present := s.home.Docs[docrepo]
	if !present {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `docpack "%s" not found`, html.EscapeString(docrepo))
		return
	}

	if subpath == "" || subpath == "/" {
		subpath = docpack.IndexPage
	} else {
		// exclude leading slash
		subpath = subpath[1:]
	}

	objpath := filepath.Join(docrepo, subpath)

	var gitobj GitObject

	obj, found := s.hotGitObjs.Get(objpath)
	if found {
		gitobj, found = obj.(GitObject)
		if !found {
			panic("not an instance of GitObject")
		}

		etag := r.Header.Get("If-None-Match")

		// etag should be a 40-byte hash enclosed by `"`.
		if len(etag) == 42 && etag[1:41] == gitobj.Hash {
			log.Println("http.StatusNotModified for Etag", etag)
			w.WriteHeader(http.StatusNotModified)
			return
		}
	} else {
		gitobj, found = GitGetHashByPath(docrepo, "HEAD", subpath)
		if !found {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `git object "%s" not found`, html.EscapeString(subpath))
			return
		}
		s.hotGitObjs.Set(objpath, gitobj, cache.DefaultExpiration)
	}

	content, ok := GitGetBlobContent(docrepo, gitobj)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `hash content "%s" not found`, html.EscapeString(gitobj.Hash))
		return
	}

	reader := bytes.NewReader(content)
	w.Header().Set("Etag", `"`+gitobj.Hash+`"`)
	w.Header().Set("Cache-Control", "private, max-age=86400") // cache +1d
	http.ServeContent(w, r, subpath, time.Time{}, reader)
}

func (s *DocityServer) repoHandler(w http.ResponseWriter, r *http.Request) {
	gitdir := filepath.Join(s.home.Path, "repos", "filemaker16en.git")
	gitserver := NewGitServer(gitdir)
	gitserver.ServeHTTP(w, r)
}
