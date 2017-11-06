package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	ErrorDirIsNotGitRepo = iota
	ErrorDirIsNotBriefcase
	ErrorBriefcaseHasNoPackageName
	ErrorBriefcaseHasNoObjectsBase
)

const (
	DefaultBriefcaseHome = "$HOME/.gitbriefcase"
	BriefcaseConfigInGit = "briefcase/config"
)

func checkGitVersion() {
}

func checkBriefcaseHome() {
	briefcaseHome := DefaultBriefcaseHome

	cmd := exec.Command("git", "config", "--global", "--get", "briefcase.home")
	if out, err := cmd.Output(); err == nil {
		briefcaseHome = strings.TrimSpace(string(out))
	}

	briefcaseHome = os.ExpandEnv(briefcaseHome)
	if !filepath.IsAbs(briefcaseHome) {
		log.Fatalf("briefcase home \"%s\" is not an absolute path", briefcaseHome)
	}

	if fileInfo, err := os.Stat(briefcaseHome); err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("briefcase home \"%s\" doesn't exist", briefcaseHome)
		} else {
			log.Fatal(err)
		}
	} else {
		if !fileInfo.IsDir() {
			log.Fatalf("\"%s\" is not a directory", briefcaseHome)
		}
	}

	if err := os.Chdir(briefcaseHome); err != nil {
		log.Fatal(err)
	}

	if gitdirs, err := filepath.Glob("*.git"); err != nil {
		// The only possible returned error is ErrBadPattern, when pattern is malformed.
		log.Fatal(err)
	} else {
		for _, gitdir := range gitdirs {
			checkBriefcaseConfig(gitdir)
		}
	}
}

// XXX: should not fatal for individual directory
// XXX: should use some exeception handling
func checkBriefcaseConfig(gitdir string) {
	config := filepath.Join(gitdir, BriefcaseConfigInGit)
	if fileInfo, err := os.Stat(config); err != nil {
		// opportunity for refactoring
		if os.IsNotExist(err) {
			log.Fatalf("briefcase config \"%s\" doesn't exist", config)
		} else {
			log.Fatal(err)
		}
	} else {
		if !fileInfo.Mode().IsRegular() {
			log.Fatalf("\"%s\" is not a regular file", config)
		}
	}

	if out, err := exec.Command("git", "config", "-f", config, "-l").Output(); err != nil {
		log.Fatal(err)
	} else {
		parseBriefcaseConfig(string(out))
	}
}

func parseBriefcaseConfig(config string) map[string]string {
	items := make(map[string]string)
	re := regexp.MustCompile(`(?m:^briefcase\.(.+)=(.+)$)`)
	for _, matches := range re.FindAllStringSubmatch(config, -1) {
		key, value := matches[1], matches[2]
		items[key] = value
	}
	return items
}

func main() {
	out, err := exec.Command("git", "config", "--get", "user.name").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(out))

	checkGitVersion()
	checkBriefcaseHome()
}
