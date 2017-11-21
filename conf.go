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
	DefaultBriefcaseHome        = "$HOME/.gitbriefcase"
	PathBriefcaseConfigInGitDir = "briefcase/config"
)

// ConfigItem contains parameters of a briefcase
type ConfigItem struct {
	subsectionName string
	displayName    string
	objectsBase    string
	indexPage      string
}

// ConfigItemDict ...
type ConfigItemDict map[string]*ConfigItem

func changeToBriefcaseHome() {
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

func scanGitRepos() (bfcList []*Briefcase) {
	if gitdirs, err := filepath.Glob("*.git"); err != nil {
		// The only possible returned error is ErrBadPattern, when pattern is malformed.
		log.Fatal(err)
	} else {
		for _, gitdir := range gitdirs {
			if config, ok := readConfig(gitdir); !ok {
				continue
			} else {
				itemDict := make(ConfigItemDict)
				parseConfig(config, itemDict)
				checkConfig(itemDict)
				// TODO: make a list of briefcase
			}
		}
	}
	return
}

func readConfig(gitdir string) (string, bool) {
	configFile := filepath.Join(gitdir, PathBriefcaseConfigInGitDir)

	if fileInfo, err := os.Stat(configFile); err != nil {
		if os.IsNotExist(err) {
			log.Printf("briefcase config \"%s\" doesn't exist\n", configFile)
		} else {
			log.Println(err)
		}
		return "", false
	} else if !fileInfo.Mode().IsRegular() {
		log.Fatalf("\"%s\" is not a regular file", configFile)
		return "", false
	} else {
		if out, err := exec.Command("git", "config", "-f", configFile, "-l").Output(); err != nil {
			log.Println(err)
			return "", false
		} else {
			return string(out), false
		}
	}
}

func parseConfig(config string, itemDict ConfigItemDict) {

	re := regexp.MustCompile(`(?m:^briefcase(\..+)?(\..+)=(.+)$)`)

	for _, matches := range re.FindAllStringSubmatch(config, -1) {
		subsection, key, value := matches[1], matches[2], matches[3]
		if subsection == "" {
			// \n is not allowed in subsection name in git-config
			// it is hereby used to denote empty subsection
			subsection = "\n"
		} else {
			subsection = subsection[1:] // exclude the leading '.'
		}
		configItem, ok := itemDict[subsection]
		if !ok {
			configItem = &ConfigItem{}
			configItem.subsectionName = subsection
			itemDict[subsection] = configItem
		}
		switch key {
		case ".displayname":
			configItem.displayName = value
		case ".objectsbase":
			configItem.objectsBase = value
		case ".indexpage":
			configItem.indexPage = value
		default:
			log.Printf("unrecognized configuration \"%s\"\n", matches[0])
		}
	}
}

func checkConfig(itemDict ConfigItemDict) {
	for name, item := range itemDict {
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
			delete(itemDict, name)
			log.Printf("removing %s from the configuration", name)
		}
	}
}
