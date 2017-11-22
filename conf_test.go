package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const configSample = `
[briefcase]
displayName = Python 3.6
objectsBase = HEAD
indexPage = index.html
[briefcase "hello"]
indexPage = open.html`

const outputSample = `briefcase.displayname=Python 3.6
briefcase.objectsbase=HEAD
briefcase.indexpage=index.html
briefcase.hello.indexpage=open.html
`

func TestParseConfig(t *testing.T) {
	itemdict := make(DocItemMap)
	parseConfig(outputSample, itemdict)

	compString := func(got, expected string, description string) {
		if got != expected {
			t.Error(
				description,
				"expected", expected,
				"got", got,
			)
		}
	}
	if item, ok := itemdict["\n"]; !ok {
		t.Error("default section not found")
	} else {
		compString(item.subsectionName, "\n", "\\n.subsectionName")
		compString(item.displayName, "Python 3.6", "\\n.displayName")
		compString(item.objectsBase, "HEAD", "\\n.objectsBase")
		compString(item.indexPage, "index.html", "\\n.indexPage")
	}
	if item, ok := itemdict["hello"]; !ok {
		t.Error("hello subsection not found")
	} else {
		compString(item.indexPage, "open.html", "hello.indexPage")
	}

	checkConfig(itemdict)
	if _, ok := itemdict["hello"]; ok {
		t.Error("hello subsection should have been removed")
	}
}

func MustGit(args ...string) string {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(out))
}

func MustTempDir(dir, prefix string) string {
	dir, err := ioutil.TempDir(dir, prefix)
	if err != nil {
		panic(err)
	}
	return dir
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func TestReadConfig(t *testing.T) {
	theRootDir := MustGit("rev-parse", "--show-toplevel")
	theGitDir := filepath.Join(theRootDir, ".git")
	tempDir := MustTempDir(theGitDir, "test")
	defer os.RemoveAll(tempDir)

	confFile := filepath.Join(tempDir, "config")
	Must(ioutil.WriteFile(confFile, []byte(configSample), 0666))

	GitBriefcaseConfDir = filepath.Base(tempDir)
	config, ok := readConfig(theGitDir)
	if !ok {
		t.Fatal("failed to read config in", tempDir)
	}
	if config != outputSample {
		t.Fatalf("expected:\n %#v\ngot:\n%#v\n", outputSample, config)
	}
}
