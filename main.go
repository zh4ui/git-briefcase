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

// XXX
const (
	ErrDirIsNotGitRepo = iota
	ErrDirIsNotBriefcase
	ErrBriefcaseHasNoPackageName
	ErrBriefcaseHasNoObjectsBase
)

const (
	DefaultBriefcaseHome          = "$HOME/.gitbriefcase"
	PathOfBriefcaseConfigInGitDir = "briefcase/config"
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

func changeToBriefcaseHome() {
	bfcHome := DefaultBriefcaseHome

	cmd := exec.Command("git", "config", "--global", "--get", "briefcase.home")
	if out, err := cmd.Output(); err == nil {
		bfcHome = strings.TrimSpace(string(out))
	}

	bfcHome = os.ExpandEnv(bfcHome)
	if !filepath.IsAbs(bfcHome) {
		log.Fatalf("briefcase shop \"%s\" is not an absolute path", bfcHome)
	}

	if fileInfo, err := os.Stat(bfcHome); err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("briefcase shop \"%s\" doesn't exist", bfcHome)
		} else {
			log.Fatal(err)
		}
	} else {
		if !fileInfo.IsDir() {
			log.Fatalf("\"%s\" is not a directory", bfcHome)
		}
	}

	if err := os.Chdir(bfcHome); err != nil {
		log.Fatal(err)
	}
}

func checkBriefcases() (bfcList []*Briefcase) {

	changeToBriefcaseHome()

	if gitdirs, err := filepath.Glob("*.git"); err != nil {
		// The only possible returned error is ErrBadPattern, when pattern is malformed.
		log.Fatal(err)
	} else {
		for _, gitdir := range gitdirs {
			conf, ok := readConfig(gitdir)
			newBfcList := parseConfig(conf)
			bfcList = append(newBfcList)
			bfc := NewBriefcase(gitdir)
			if checkBriefcaseConfig(bfc) {
				bfcList = append(bfcList, bfc)
			}
		}
	}
	return
}

func readConfig(gitdir string) (string, bool) {
	configFile := filepath.Join(gitdir, PathOfBriefcaseConfigInGitDir)

	if fileInfo, err := os.Stat(configFile); err != nil {
		if os.IsNotExist(err) {
			log.Printf("briefcase config \"%s\" doesn't exist\n", configFile)
		} else {
			log.Println(err)
		}
		return "", false
	} else if !fileInfo.Mode().IsRegular() {
		log.Fatalf("\"%s\" is not a regular file", configFile)
		return "", false
	} else {
		if out, err := exec.Command("git", "config", "-f", configFile, "-l").Output(); err != nil {
			log.Println(err)
			return "", false
		} else {
			return string(out), false
		}
	}
}

// ConfigItem ...
type ConfigItem struct {
	subsection, name, value string
}

func parseConfig(config string) {
	var items []ConfigItem
	re := regexp.MustCompile(`(?m:^briefcase(\..+)?(\..+)=(.+)$)`)
	for _, matches := range re.FindAllStringSubmatch(config, -1) {
		subsection, name, value := matches[1], matches[2], matches[3]
		if subsection != "" {
			subsection = subsection[1:] // exclude the leading '.'
		}
		items = append(items, ConfigItem{subsection, name, value})
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

	bfcList := checkBriefcases()

	_ = bfcList

	http.Handle("/", http.HandlerFunc(bfcIndex))
	err := http.ListenAndServe(*httpAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
