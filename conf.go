package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	GitDocityDefaultHome = "$HOME/.gitdocity"
	GitDocityReposDir    = "repos"
)

var GitDocityHome string

type DocPack struct {
	IndexPage string
}

func (d *DocPack) GetServingUrlPath(gitdir string) string {
	// XXX need to check whether this is safe
	p := filepath.Join("doc", gitdir, d.IndexPage)
	return filepath.ToSlash(filepath.Clean(p))
}

type GitDocity struct {
	Home string
	// docs are organized using map, and indexed by name
	Docs        map[string]*DocPack
	InvalidDocs map[string][]string
}

func NewGitDocity() *GitDocity {
	g := &GitDocity{}
	g.Docs = make(map[string]*DocPack)
	g.InvalidDocs = make(map[string][]string)
	g.init()
	return g
}

func (g *GitDocity) init() {
	g.getHome()
	g.readDocPacks()
}

func (g *GitDocity) getHome() {
	cmd := exec.Command("git", "config", "--global", "--get", "docity.home")
	out, err := cmd.Output()
	if err != nil {
		log.Println(err)
	}

	g.Home = strings.TrimSpace(string(out))
	if g.Home != "" {
		log.Print("git-docity home is globally configured as:", g.Home)
	} else {
		log.Print("git-docity home is set to default:", g.Home)
	}

	g.Home = os.ExpandEnv(g.Home)

	if !filepath.IsAbs(g.Home) {
		log.Fatalf("git-docity home \"%s\" is not an absolute path", g.Home)
	}

	if fileInfo, err := os.Stat(g.Home); err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("git-docity home \"%s\" doesn't exist", g.Home)
		} else {
			log.Fatal(g.Home, err)
		}
	} else {
		if !fileInfo.IsDir() {
			log.Fatalf("git-docity home \"%s\" is not a directory", g.Home)
		}
	}

	if err := os.Chdir(g.Home); err != nil {
		log.Fatal(err)
	}
}

func (g *GitDocity) readDocPacks() {

	// Should be in GitDocity Home
	jsonfiles, err := filepath.Glob("*.json")

	if err != nil {
		// the only err is ErrBadPattern, which is assumed as a programming error.
		// panic() is used to exit the program.
		panic(err)
	}

	for _, jsonfile := range jsonfiles {

		var errors []string

		docname := strings.TrimSuffix(jsonfile, ".json")

		jsonbytes, err := ioutil.ReadFile(jsonfile)
		if err != nil {
			errors = append(errors, err.Error())
			g.InvalidDocs[docname] = errors
			continue
		}

		docpack := &DocPack{}
		err = json.Unmarshal(jsonbytes, docpack)
		if err != nil {
			errors = append(errors, err.Error())
			g.InvalidDocs[docname] = errors
			continue
		}

		errors = checkDocPack(docname, docpack)
		if len(errors) != 0 {
			g.InvalidDocs[docname] = errors
		} else {
			g.Docs[docname] = docpack
		}
	}
}

func checkDocPack(docname string, docpack *DocPack) (errors []string) {

	gitdir := filepath.Join(GitDocityReposDir, docname+".git")

	if !GitIsRepo(gitdir) {
		problem := fmt.Sprintf("not a valid git repo: \"%s\"\n", gitdir)
		errors = append(errors, problem)
	}

	if docpack.IndexPage == "" {
		problem := fmt.Sprintf("indexPage not specified for \"%s\"\n", gitdir)
		errors = append(errors, problem)
	}

	return
}
