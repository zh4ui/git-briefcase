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
	indexPage string
}

type GitBriefcase struct {
	home string
	docs map[string]*DocPack // map path to *DocPack
}

func (g *GitBriefcase) Init() {
	g.findHome()
	if err := os.Chdir(g.home); err != nil {
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

	g.home = strings.TrimSpace(string(out))
	if g.home != "" {
		log.Println("git-breifcase home is configured as:", g.home)
	} else {
		g.home = os.ExpandEnv(GitBriefcaseDefaultHome)
		log.Println("git-breifcase home is set to default:", g.home)
	}

	if !filepath.IsAbs(g.home) {
		log.Fatalf("git-briefcase home \"%s\" is not an absolute path\n", g.home)
	}

	if fileInfo, err := os.Stat(g.home); err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("git-briefcase home \"%s\" doesn't exist", g.home)
		} else {
			log.Fatal(g.home, err)
		}
	} else {
		if !fileInfo.IsDir() {
			log.Fatalf("git-briefcase home \"%s\" is not a directory", g.home)
		}
	}
}

func (g *GitBriefcase) readConfig() string {

	configPath := filepath.Join(g.home, GitBriefcaseConfigFile)

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
		log.Fatalln(configPath, err)
	}

	return string(out)
}

func (g *GitBriefcase) parseConfig(config string) {

	re := regexp.MustCompile(`(?m:^docpack\.(.+)\.(.+)=(.+)$)`)

	g.docs = make(map[string]*DocPack)

	for _, matches := range re.FindAllStringSubmatch(config, -1) {
		subsection, param, value := matches[1], matches[2], matches[3]
		docpack := g.docs[subsection]
		if docpack == nil {
			docpack = &DocPack{}
			g.docs[subsection] = docpack
		}
		switch param {
		case "indexpage":
			docpack.indexPage = value
		default:
			log.Printf("unrecognized configuration \"%s\"\n", matches[0])
		}
	}
}

func (g *GitBriefcase) checkDocPacks() {
	for gitdir, docpack := range g.docs {
		ok := true
		if docpack.indexPage == "" {
			log.Printf("indexPage not specified for \"%s\"\n", gitdir)
			ok = false
		}
		if !isGitRepo(gitdir) {
			log.Printf("not a valid git repo: \"%s\"\n", gitdir)
			ok = false
		}
		if !ok {
			log.Printf("removing invalid DocPack \"%s\"\n", gitdir)
			delete(g.docs, gitdir)
		}
	}
}
