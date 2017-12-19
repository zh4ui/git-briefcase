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

type GitDocityServer struct {
	docity     *GitDocity
	tmpl       *template.Template
	staticDir  string
	hotGitObjs *cache.Cache
}

func NewGitDocityServer(staticDir string) *GitDocityServer {
	s := &GitDocityServer{}
	s.staticDir = staticDir
	s.docity = NewGitDocity()
	s.tmpl = template.New("git-docity")
	// 5 minutes expiration and 60 minutes purge period
	s.hotGitObjs = cache.New(10*time.Minute, 60*time.Minute)

	assetsDir := filepath.Join(s.staticDir, "assets")
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))))

	http.HandleFunc("/docpack/", s.docpackHandler)
	http.HandleFunc("/git/", s.gitHandler)
	http.HandleFunc("/", s.rootHandler)
	// TODO: serve static file here

	return s
}

func (s *GitDocityServer) Run(servingAddr string) {
	templatesDir := filepath.Join(s.staticDir, "templates")
	indexPage := filepath.Join(templatesDir, "index.gohtml")
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
	s.tmpl.ExecuteTemplate(w, "index.gohtml", s.docity)
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
	docpack, present := s.docity.Docs[docname]
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

	objpath := filepath.Join(gitdir, subpath)
	obj, found := s.hotGitObjs.Get(objpath)

	var gitobj GitObject

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
		gitobj, found = GitGetHashByPath(gitdir, "HEAD", subpath)
		if !found {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `git object "%s" not found`, html.EscapeString(subpath))
			return
		}
		s.hotGitObjs.Set(objpath, gitobj, cache.DefaultExpiration)
	}

	content, ok := GitGetBlobContent(gitdir, gitobj.Hash)
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

func (s *GitDocityServer) gitHandler(w http.ResponseWriter, r *http.Request) {
	gitdir := filepath.Join(s.docity.Home, "repos", "filemaker16en.git")
	gitserver := NewGitServer(gitdir)
	gitserver.ServeHTTP(w, r)
}

func (s *GitDocityServer) errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprint(w, "custom 404")
	}
}
