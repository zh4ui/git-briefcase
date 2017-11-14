package main

import "testing"

func TestParseBriefcaseConfig(t *testing.T) {
	const config = `
briefcase.displayname=Python 3.6
briefcase.objectsbase=HEAD
briefcase.indexpage=index.html`
	items := parseBriefcaseConfig(config)
	kvtest := func(k, v string) {
		if item, ok := items[k]; ok {
			if item != v {
				t.Error(
					k,
					"expected", v,
					"got", item,
				)
			}
		} else {
			t.Error(k, "not found")
		}
	}
	kvtest("displayname", "Python 3.6")
	kvtest("objectsbase", "HEAD")
	kvtest("indexpage", "index.html")
}
