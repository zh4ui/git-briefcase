package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

const (
	DefaultBriefcaseHome = "$HOME/.gitbriefcase"
)

var (
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
