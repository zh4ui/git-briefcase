package main

import "testing"

func TestParseConfig(t *testing.T) {
	const config = `
briefcase.displayname=Python 3.6
briefcase.objectsbase=HEAD
briefcase.indexpage=index.html
briefcase.hello.indexpage=open.html`

	itemdict := make(ConfigItemDict)
	parseConfig(config, itemdict)

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
