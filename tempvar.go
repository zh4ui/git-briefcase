// data structure for tempalte use
package main

import (
	"fmt"
	"path/filepath"
)

const DefaultIndexPage = "index.html"

type DocPack struct {
	Description string
}

func (d *DocPack) GetViewUrlPath(gitdir string) string {
	// XXX need to check whether this is safe
	p := filepath.Join("view", gitdir, DefaultIndexPage)
	return filepath.ToSlash(filepath.Clean(p))
}

func (d *DocPack) GetRepoUrlPath(gitdir string) string {
	// XXX need to check whether this is safe
	p := filepath.Join("repo", gitdir)
	return filepath.ToSlash(filepath.Clean(p))
}

func (d *DocPack) GetConfUrlPath(gitdir string) string {
	// XXX need to check whether this is safe
	p := filepath.Join("conf", gitdir)
	return filepath.ToSlash(filepath.Clean(p))
}

func (d *DocPack) GetGitwebUrlPath(gitdir string) string {
	// XXX need to check whether this is safe
	return fmt.Sprintf("/gitweb/?p=%s", gitdir)
}
