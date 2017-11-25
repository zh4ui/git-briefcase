package main

import (
	"log"
	"os/exec"
)

func checkGitVersion() {
	if out, err := exec.Command("git", "--version").Output(); err != nil {
		log.Fatal(err)
	} else {
		// currently no use is made out of the output
		_ = out
	}
}

func isGitRepo(gitdir string) bool {
	_, err := exec.Command("git", "--git-dir", gitdir, "rev-parse").Output()
	if err != nil {
		return false
	} else {
		return true
	}
}
