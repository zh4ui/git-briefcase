package main

import "testing"

func TestParseBriefcaseConfig(t *testing.T) {
	const config = `
briefcase.displayname=Python 3.6
briefcase.objectsbase=HEAD
briefcase.indexpage=index.html
briefcase.hello.indexpage=open.html`

	bfc := NewBriefcase("")
	parseBriefcaseConfig(config, bfc)
	if _, ok := bfc.docs[".\n"]; !ok {
		t.Error("default section not found")
	}
	if _, ok := bfc.docs[".hello"]; !ok {
		t.Error("hello subsection not found")
	}
	testParams := func(k, v string) {
		if param, ok := bfc.params[k]; ok {
			if param != v {
				t.Error(
					k,
					"expected", v,
					"got", param,
				)
			}
		} else {
			t.Error(k, "not found")
		}
	}
	testParams(".\n.displayname", "Python 3.6")
	testParams(".\n.objectsbase", "HEAD")
	testParams(".\n.indexpage", "index.html")
	testParams(".hello.indexpage", "open.html")
}
