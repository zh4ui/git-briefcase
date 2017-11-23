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
	urlPath   string
	indexPage string
}

type GitBriefcase struct {
	home string
	docs []*DocPack
}

func (g *GitBriefcase) init() {

	cmd := exec.Command("git", "config", "--global", "--get", "briefcase.home")
	out, err := cmd.Output()
	if err != nil {
		log.Println(err)
	}

	g.home = strings.TrimSpace(string(out))
	if g.home != "" {
		log.Println("git-breifcase home is configured as:", g.home)
	} else {
		g.home = os.ExpandEnv(GitBriefcaseDefaultDir)
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
		if !fileInfo.IsDir(g.home) {
			log.Fatalf("git-briefcase home \"%s\" is not a directory", g.home)
		}
	}
}

func (g *GitBriefcase) readConfig() {

	configPath := filepath.Join(g.home, GitBriefcaseConfigFile)

	if fileInfo, err := os.Stat(configPath); err != nil {
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
	if out, err := cmd.Output(); err != nil {
		log.Fatalln(configPath, err)
	}

	config := string(out)
}

func (g *GitBriefcase) parseConfig() {
	re := regexp.MustCompile(`(?m:^docpack\.(.+)\.(.+)=(.+)$)`)

	for _, matches := range re.FindAllStringSubmatch(config, -1) {
		docpack, param, value := matches[1], matches[2], matches[3]
		// XXX: TO RESUME
		docItem, ok := itemMap[subsection]
		if !ok {
			docItem = &DocItem{}
			docItem.subsectionName = subsection
			itemMap[subsection] = docItem
		}
		switch key {
		case ".displayname":
			docItem.displayName = value
		case ".objectsbase":
			docItem.objectsBase = value
		case ".indexpage":
			docItem.indexPage = value
		default:
			log.Printf("unrecognized configuration \"%s\"\n", matches[0])
		}
	}
}

func scanGitRepos() (allDocs []DocFeed) {
	if gitdirs, err := filepath.Glob("*.git"); err != nil {
		// The only possible returned error is ErrBadPattern, when pattern is malformed.
		log.Fatal(err)
	} else {
		for _, gitdir := range gitdirs {
			if config, ok := readConfig(gitdir); !ok {
				continue
			} else {
				itemMap := make(DocItemMap)
				parseConfig(config, itemMap)
				checkConfig(itemMap)
				docFeed := DocFeed{}
				docFeed.gitdir = gitdir
				docFeed.items = make([]*DocItem, len(itemMap))
				for name := range itemMap {
					docFeed.items = append(docFeed.items, itemMap[name])
				}
				allDocs = append(allDocs, docFeed)
			}
		}
	}
	return
}

/*
func changeToBriefcaseHomeDir() {
	if err := os.Chdir(bfcHome); err != nil {
		log.Fatal(err)
	}
}
*/
