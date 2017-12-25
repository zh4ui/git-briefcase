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
