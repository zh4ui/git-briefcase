package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	DocityDefaultHome    = "$HOME/.gitdocity"
	DocityConfigFileName = "docity.json"
)

type DocityHome struct {
	Path string
	Docs map[string]*DocPack
}

func NewDocityHome() *DocityHome {
	g := &DocityHome{}
	g.Docs = make(map[string]*DocPack)
	g.findHome()
	g.readDocs()
	return g
}

func (g *DocityHome) findHome() {
	g.Path = GitConfigGetHome()
	if g.Path != "" {
		log.Print("GitDocity: home is globally configured as: ", g.Path)
	} else {
		log.Print("GitDocity: home is set to default: ", g.Path)
	}
	g.Path = os.ExpandEnv(g.Path)

	checkDir := func(name, dir string) {
		if err := statAbsDir(dir); err != nil {
			log.Fatalf(`GitDocity: invalid %s "%s" {%s}`, name, dir, err)
		}
	}
	checkDir("home", g.Path)

	if err := os.Chdir(g.Path); err != nil {
		log.Fatalf(`GitDocity: failed to change directory to "%s" {%s}`, g.Path, err)
	}
}

func (g *DocityHome) readDocs() {
	f, err := os.Open(g.Path)
	if err != nil {
		log.Fatalf(`GitDocity: failed to open dir "%s" {%s}`, g.Path, err)
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Fatalf(`GitDocity: failed to read dir "%s" {%s}`, g.Path, err)
	}

	// read dirs of pattern "*.git"
	var repoInfos []os.FileInfo
	for _, entry := range list {
		if !entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".git") {
			continue
		}
		repoInfos = append(repoInfos, entry)
	}

	// sort by modtime
	if len(repoInfos) > 0 {
		sort.Slice(repoInfos, func(i, j int) bool {
			a := repoInfos[i].ModTime()
			b := repoInfos[j].ModTime()
			return a.Before(b)
		})
	}

	// read configs
	for _, repoInfo := range repoInfos {
		configFile := filepath.Join(repoInfo.Name(), DocityConfigFileName)
		bytes, err := ioutil.ReadFile(configFile)
		if err != nil {
			// XXX: should process error info
			continue
		}

		docpack := &DocPack{}
		err = json.Unmarshal(bytes, docpack)
		if err != nil {
			// XXX: should process error info
			continue
		}

		errors := checkDocPack(repoInfo.Name(), docpack)
		if len(errors) != 0 {
			// XXX: should process error info
			continue
		} else {
			g.Docs[repoInfo.Name()] = docpack
		}
	}
}

func checkDocPack(gitdir string, docpack *DocPack) (errors []string) {

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
