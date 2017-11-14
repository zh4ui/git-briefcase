package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	ErrDirIsNotGitRepo = iota
	ErrDirIsNotBriefcase
	ErrBriefcaseHasNoPackageName
	ErrBriefcaseHasNoObjectsBase
)

const (
	DefaultBriefcaseHome = "$HOME/.gitbriefcase"
	BriefcaseConfigInGit = "briefcase/config"
)

// should abstract the invocation of git into a function or so
func GitCmd(args ...string) {
}

func checkGitVersion() {
	if out, err := exec.Command("git", "--version").Output(); err != nil {
		log.Fatal(err)
	} else {
		// currently no use is made out of the output
		_ = out
	}
}

func checkBriefcaseShop() {
	briefcaseHome := DefaultBriefcaseHome

	cmd := exec.Command("git", "config", "--global", "--get", "briefcase.shop")
	if out, err := cmd.Output(); err == nil {
		briefcaseHome = strings.TrimSpace(string(out))
	}

	briefcaseHome = os.ExpandEnv(briefcaseHome)
	if !filepath.IsAbs(briefcaseHome) {
		log.Fatalf("briefcase shop \"%s\" is not an absolute path", briefcaseHome)
	}

	if fileInfo, err := os.Stat(briefcaseHome); err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("briefcase shop \"%s\" doesn't exist", briefcaseHome)
		} else {
			log.Fatal(err)
		}
	} else {
		if !fileInfo.IsDir() {
			log.Fatalf("\"%s\" is not a directory", briefcaseHome)
		}
	}

	if err := os.Chdir(briefcaseHome); err != nil {
		log.Fatal(err)
	}

	if gitdirs, err := filepath.Glob("*.git"); err != nil {
		// The only possible returned error is ErrBadPattern, when pattern is malformed.
		log.Fatal(err)
	} else {
		for _, gitdir := range gitdirs {
			checkBriefcaseConfig(gitdir)
		}
	}
}

type Briefcase struct {
	gitdir string
	params map[string]string
}

var g_briefcases []Briefcase

// XXX: should not fatal for individual directory
// XXX: should use some exeception handling
func checkBriefcaseConfig(gitdir string) {
	config := filepath.Join(gitdir, BriefcaseConfigInGit)
	if fileInfo, err := os.Stat(config); err != nil {
		// XXX: opportunity for refactoring
		if os.IsNotExist(err) {
			log.Fatalf("briefcase config \"%s\" doesn't exist", config)
		} else {
			log.Fatal(err)
		}
	} else {
		if !fileInfo.Mode().IsRegular() {
			log.Fatalf("\"%s\" is not a regular file", config)
		}
	}

	if out, err := exec.Command("git", "config", "-f", config, "-l").Output(); err != nil {
		log.Fatal(err)
	} else {
		params := parseBriefcaseConfig(string(out))
		g_briefcases = append(g_briefcases, Briefcase{gitdir, params})
	}
}

func parseBriefcaseConfig(config string) map[string]string {
	items := make(map[string]string)
	re := regexp.MustCompile(`(?m:^briefcase\.(.+)=(.+)$)`)
	for _, matches := range re.FindAllStringSubmatch(config, -1) {
		key, value := matches[1], matches[2]
		items[key] = value
	}
	return items
}

var tmpl = template.New("git-briefcase")

//var tmpl = template.Must(template.New("shop").Parse(templateStr))

func bfcIndex(w http.ResponseWriter, req *http.Request) {
	tmpl.ExecuteTemplate(w, "index.html", "hello!")
}

var (
	httpAddr    = flag.String("http", ":9899", "http service address") // b=98, c=99
	templateDir = flag.String("templates", "", "load templates and other web resources from this directory")
)

func handleFlags() {
	flag.Parse()

	if *templateDir != "" {
		if abspath, err := filepath.Abs(*templateDir); err != nil {
			log.Fatal(err)
		} else {
			*templateDir = abspath
			log.Printf("Using templateDir: %s\n", abspath)
		}
	}

	indexPage := filepath.Join(*templateDir, "index.html")
	tmpl = template.Must(tmpl.ParseFiles(indexPage))
}

func main() {
	handleFlags()

	checkGitVersion()
	checkBriefcaseShop()

	http.Handle("/", http.HandlerFunc(bfcIndex))
	err := http.ListenAndServe(*httpAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
