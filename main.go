package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
)

// XXX
const (
	ErrDirIsNotGitRepo = iota
	ErrDirIsNotBriefcase
	ErrBriefcaseHasNoPackageName
	ErrBriefcaseHasNoObjectsBase
)

// Briefcase represents a git repo that is configured as a breifcase and cont
type Briefcase struct {
	gitdir string
	params map[string]string
}

// NewBriefcase ...
func NewBriefcase(gitdir string) *Briefcase {
	bfc := &Briefcase{}
	bfc.gitdir = gitdir
	bfc.params = make(map[string]string, 50)
	return bfc
}

func checkGitVersion() {
	if out, err := exec.Command("git", "--version").Output(); err != nil {
		log.Fatal(err)
	} else {
		// currently no use is made out of the output
		_ = out
	}
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
	checkGitVersion()
	handleFlags()
	changeToBriefcaseHome()
	scanGitRepos()

	http.Handle("/", http.HandlerFunc(bfcIndex))
	err := http.ListenAndServe(*httpAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
