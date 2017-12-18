package main

import (
	"log"
	"os/exec"
	"regexp"
)

type GitObject struct {
	Mode string
	Type string
	Hash string
	Size string
	Name string
}

func GitLookPath() string {
	git, err := exec.LookPath("git")
	// XXX: should find an elegant error handling solution
	if err != nil {
		panic(err)
	}
	return git
}

func GitCheckVersion() {
	if out, err := exec.Command("git", "--version").Output(); err != nil {
		log.Fatal(err)
	} else {
		// currently no use is made out of the output
		_ = out
	}
}

func GitIsRepo(gitdir string) bool {
	_, err := exec.Command("git", "--git-dir", gitdir, "rev-parse").Output()
	if err != nil {
		return false
	} else {
		return true
	}
}

func GitGetHashByPath(gitdir string, treeish string, pathname string) (GitObject, bool) {
	obj := GitObject{}

	args := []string{
		"--git-dir",
		gitdir,
		"ls-tree",
		"-l",
		"HEAD",
		pathname,
	}

	cmd := exec.Command("git", args...)

	log.Println(cmd.Path, cmd.Args)

	out, err := cmd.Output()
	if err != nil {
		return obj, false
	}
	// `100644 blob 6a1f0e1014bc39d6fd9ee9b59df90f076dff00b8     119	Open Help.html`
	pattern := regexp.MustCompile(`(\S+)\s+(\S+)\s+(\S+)\s+(\d+)\s+(.+)`)
	matches := pattern.FindStringSubmatch(string(out))
	if matches == nil {
		return obj, false
	}

	obj.Mode = matches[1]
	obj.Type = matches[2]
	obj.Hash = matches[3]
	obj.Size = matches[4]
	obj.Name = matches[5]

	return obj, true
}

func GitGetBlobContent(gitdir string, hash string) ([]byte, bool) {
	args := []string{
		"--git-dir",
		gitdir,
		"cat-file",
		"blob",
		hash,
	}

	cmd := exec.Command("git", args...)

	out, err := cmd.Output()
	if err != nil {
		return nil, false
	}
	return out, true
}
