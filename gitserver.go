package main

import (
	"log"
	"net/http"
	"net/http/cgi"
	"path"
)

type GitServer struct {
	http.Handler
	gitpath string
	gitdir  string
}

func NewGitServer(gitdir string) *GitServer {
	s := &GitServer{}
	s.gitpath = GitLookPath()
	// s.gitpath = "/usr/bin/env"
	s.gitdir = gitdir

	return s
}

func (s *GitServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathTranslated := path.Join(s.gitdir, r.URL.Path[len("/git/"):])

	log.Print(pathTranslated)
	env := []string{
		"GIT_HTTP_EXPORT_ALL=",
		"PATH_TRANSLATED=" + pathTranslated,
	}

	log.Print(r.URL.Path)

	cgiHandler := cgi.Handler{
		Path: s.gitpath,
		Args: []string{"http-backend"},
		Env:  env,
	}
	cgiHandler.ServeHTTP(w, r)
}
