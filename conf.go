package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	GitBriefcaseDefaultHome = "$HOME/.gitbriefcase"
	GitBriefcaseConfigFile  = "docpacks"
)

type DocPack struct {
	IndexPage string
}

type GitBriefcase struct {
	Home string
	Docs map[string]*DocPack // map path to *DocPack
}

func NewGitBriefcase() *GitBriefcase {
	gb := &GitBriefcase{}
	gb.init()
	return gb
}

func (g *GitBriefcase) init() {
	g.findHome()
	if err := os.Chdir(g.Home); err != nil {
		log.Fatal(err)
	}
	config := g.readConfig()
	g.parseConfig(config)
	g.checkDocPacks()
}

func (g *GitBriefcase) findHome() {
	cmd := exec.Command("git", "config", "--global", "--get", "briefcase.home")
	out, err := cmd.Output()
	if err != nil {
		log.Println(err)
	}

	g.Home = strings.TrimSpace(string(out))
	if g.Home != "" {
		log.Print("git-breifcase home is globally configured as:", g.Home)
	} else {
		log.Print("git-breifcase home is set to default:", g.Home)
	}

	g.Home = os.ExpandEnv(GitBriefcaseDefaultHome)

	if !filepath.IsAbs(g.Home) {
		log.Fatalf("git-briefcase home \"%s\" is not an absolute path\n", g.Home)
	}

	if fileInfo, err := os.Stat(g.Home); err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("git-briefcase home \"%s\" doesn't exist", g.Home)
		} else {
			log.Fatal(g.Home, err)
		}
	} else {
		if !fileInfo.IsDir() {
			log.Fatalf("git-briefcase home \"%s\" is not a directory", g.Home)
		}
	}
}

func (g *GitBriefcase) readConfig() string {

	configPath := filepath.Join(g.Home, GitBriefcaseConfigFile)

	fileInfo, err := os.Stat(configPath)

	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("git-briefcase config \"%s\" doesn't exist\n", configPath)
		} else {
			log.Fatalln(configPath, err)
		}
	}

	if !fileInfo.Mode().IsRegular() {
		log.Fatalf("git-briefcase config \"%s\" is not a regular file", configPath)
	}

	cmd := exec.Command("git", "config", "-f", configPath, "-l")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(configPath, err)
	}

	return string(out)
}

func (g *GitBriefcase) parseConfig(config string) {

	re := regexp.MustCompile(`(?m:^docpack\.(.+)\.(.+)=(.+)$)`)

	g.Docs = make(map[string]*DocPack)

	for _, matches := range re.FindAllStringSubmatch(config, -1) {
		subsection, param, value := matches[1], matches[2], matches[3]
		docpack := g.Docs[subsection]
		if docpack == nil {
			docpack = &DocPack{}
			g.Docs[subsection] = docpack
		}
		switch param {
		case "indexpage":
			docpack.IndexPage = value
		default:
			log.Printf("unrecognized configuration \"%s\"\n", matches[0])
		}
	}
}

func (g *GitBriefcase) checkDocPacks() {

	for gitdir, docpack := range g.Docs {
		ok := true
		if docpack.IndexPage == "" {
			log.Printf("indexPage not specified for \"%s\"\n", gitdir)
			ok = false
		}
		if !isGitRepo(gitdir) {
			log.Printf("not a valid git repo: \"%s\"\n", gitdir)
			ok = false
		}
		if !ok {
			log.Printf("removing invalid DocPack \"%s\"\n", gitdir)
			delete(g.Docs, gitdir)
		}
	}
}
