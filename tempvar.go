// data structure for tempalte use
package main

import "path/filepath"

type DocPack struct {
	IndexPage   string
	Description string
}

func (d *DocPack) GetViewUrlPath(gitdir string) string {
	// XXX need to check whether this is safe
	p := filepath.Join("view", gitdir, d.IndexPage)
	return filepath.ToSlash(filepath.Clean(p))
}

func (d *DocPack) GetConfUrlPath(gitdir string) string {
	// XXX need to check whether this is safe
	p := filepath.Join("conf", gitdir)
	return filepath.ToSlash(filepath.Clean(p))
}
