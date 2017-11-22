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
	DefaultBriefcaseHome = "$HOME/.gitbriefcase"
)

var (
	GitBriefcaseHomeDir = DefaultBriefcaseHome
	GitBriefcaseConfDir = "briefcase"
)

// DocFeed ...
type DocFeed struct {
	gitdir string
	items  []*DocItem
}

// DocItem contains parameters of a briefcase
type DocItem struct {
	subsectionName string "optional"
	displayName    string "optional"
	objectsBase    string
	indexPage      string
}

// DocItemMap ...
type DocItemMap map[string]*DocItem

func changeToBriefcaseHomeDir() {
	bfcHome := DefaultBriefcaseHome

	cmd := exec.Command("git", "config", "--global", "--get", "briefcase.home")
	if out, err := cmd.Output(); err == nil {
		bfcHome = strings.TrimSpace(string(out))
	}

	bfcHome = os.ExpandEnv(bfcHome)
	if !filepath.IsAbs(bfcHome) {
		log.Fatalf("briefcase shop \"%s\" is not an absolute path", bfcHome)
	}

	if fileInfo, err := os.Stat(bfcHome); err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("briefcase shop \"%s\" doesn't exist", bfcHome)
		} else {
			log.Fatal(err)
		}
	} else {
		if !fileInfo.IsDir() {
			log.Fatalf("\"%s\" is not a directory", bfcHome)
		}
	}

	if err := os.Chdir(bfcHome); err != nil {
		log.Fatal(err)
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

// TODO: use panic to handle error
func readConfig(gitdir string) (string, bool) {
	configFile := filepath.Join(gitdir, GitBriefcaseConfDir, "config")

	if fileInfo, err := os.Stat(configFile); err != nil {
		if os.IsNotExist(err) {
			log.Printf("briefcase config \"%s\" doesn't exist\n", configFile)
		} else {
			log.Println(configFile, err)
		}
		return "", false
	} else if !fileInfo.Mode().IsRegular() {
		log.Fatalf("\"%s\" is not a regular file", configFile)
		return "", false
	} else {
		cmd := exec.Command("git", "config", "-f", configFile, "-l")
		if out, err := cmd.Output(); err != nil {
			log.Println(configFile, err)
			return "", false
		} else {
			return string(out), true
		}
	}
}

func parseConfig(config string, itemMap DocItemMap) {

	re := regexp.MustCompile(`(?m:^briefcase(\..+)?(\..+)=(.+)$)`)

	for _, matches := range re.FindAllStringSubmatch(config, -1) {
		subsection, key, value := matches[1], matches[2], matches[3]
		if subsection == "" {
			// \n is not allowed in subsection name in git-config
			// it is hereby used to denote empty subsection
			subsection = "\n"
			// TODO: stop doing it
		} else {
			subsection = subsection[1:] // exclude the leading '.'
		}
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

func checkConfig(itemMap DocItemMap) {
	for name, item := range itemMap {
		ok := true
		if item.objectsBase == "" {
			log.Printf("%s has no objectsBase\n", name)
			ok = false
		}
		if item.indexPage == "" {
			log.Printf("%s has no indexPage\n", name)
			ok = false
		}
		if !ok {
			delete(itemMap, name)
			log.Printf("removing %s from the configuration", name)
		}
	}
}
