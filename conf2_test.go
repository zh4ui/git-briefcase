package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

const configSample = `
[DocPack "python3.6"]
	indexPage = index.html
[DocPack "hello"]
	indexPage = welcome.html
`

func ohno(err error) {
	if err != nil {
		panic(err)
	}
}

func readConfigSample() string {
	tempf, err := ioutil.TempFile("", "docpack")
	ohno(err)
	_, err = tempf.WriteString(configSample)
	ohno(err)
	tempf.Close()
	defer os.Remove(tempf.Name())

	out, err := exec.Command("git", "config", "-f", tempf.Name(), "-l").Output()
	ohno(err)

	return string(out)
}

func compString(t *testing.T, got, expected string, description string) {
	if got != expected {
		t.Error(
			description,
			"expected", expected,
			"got", got,
		)
	}
	return
}

func TestParseConfig(t *testing.T) {
	gb := GitBriefcase{}
	config := readConfigSample()
	gb.parseConfig(config)

	if docpack, ok := gb.docs["python3.6"]; !ok {
		t.Errorf("DocPack \"%s\" not found\n", docpack)
	} else {
		compString(t, docpack.indexPage, "index.html", "docpack.indexPage")
	}

	if docpack, ok := gb.docs["hello"]; !ok {
		t.Errorf("DocPack \"%s\" not found\n", docpack)
	} else {
		compString(t, docpack.indexPage, "welcome.html", "docpack.indexPage")
	}
}
