package main

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"net/http/cgi"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"
)

type DocityServer struct {
	home       *DocityHome
	tmpl       *template.Template
	config     *DocityConfig
	staticDir  string
	hotGitObjs *cache.Cache
}

func NewDocityServer(staticDir string) *DocityServer {
	s := &DocityServer{}

	s.staticDir = staticDir
	s.config = GitGetConfig()

	s.home = NewDocityHome(s.config.Home)
	s.tmpl = template.New("docity")
	// 5 minutes expiration and 60 minutes purge period
	s.hotGitObjs = cache.New(10*time.Minute, 60*time.Minute)

	http.Handle("/assets/",
		http.StripPrefix("/assets/",
			http.FileServer(http.Dir(filepath.Join(s.staticDir, "assets")))))

	http.HandleFunc("/gitweb/", s.serveGitweb)
	http.HandleFunc("/view/", s.serveView)
	http.HandleFunc("/repo/", s.serveRepo)
	http.HandleFunc("/", s.serveRoot)

	return s
}

func (s *DocityServer) Run(servingAddr string) {
	templatesDir := filepath.Join(s.staticDir, "templates")
	indexPage := filepath.Join(templatesDir, "index.gohtml")
	s.tmpl = template.Must(s.tmpl.ParseFiles(indexPage))

	log.Print("ListenAndServe", servingAddr)
	err := http.ListenAndServe(servingAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *DocityServer) serveRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// XXX re-compile tempalte for each request,
	// should turn off for release version
	s.tmpl = template.New("docity")
	templatesDir := filepath.Join(s.staticDir, "templates")
	indexPage := filepath.Join(templatesDir, "index.gohtml")
	s.tmpl = template.Must(s.tmpl.ParseFiles(indexPage))
	s.tmpl.ExecuteTemplate(w, "index.gohtml", s.home)
}

func (s *DocityServer) serveView(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/view/" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

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

	_ = docpack

	if subpath == "" || subpath == "/" {
		subpath = DefaultIndexPage
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

func (s *DocityServer) serveRepo(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/repo/" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	pathTrans := filepath.Join(s.home.Path, r.URL.Path[len("/repo/"):])

	env := []string{
		"GIT_HTTP_EXPORT_ALL=",
		"PATH_TRANSLATED=" + pathTrans,
	}

	// TODO dump cgi error to client

	cgiHandler := cgi.Handler{
		Path: MustFindGit(),
		Args: []string{"http-backend"},
		Env:  env,
	}

	cgiHandler.ServeHTTP(w, r)
}

func (s *DocityServer) serveGitweb(w http.ResponseWriter, r *http.Request) {

	if strings.HasPrefix(r.URL.Path, "/gitweb/static/") {
		http.ServeFile(w, r, filepath.Join(`/usr/local/Cellar/git/2.15.1_1/share/`, r.URL.Path))
		return
	}

	env := []string{
		"GITWEB_CONFIG=" + filepath.Join(s.home.Path, "gitweb.conf"),
	}

	gitweb := cgi.Handler{
		Path: `/usr/local/Cellar/git/2.15.1_1/share/gitweb/gitweb.cgi`,
		Root: `/gitweb/`,
		Env:  env,
	}
	gitweb.ServeHTTP(w, r)
}
