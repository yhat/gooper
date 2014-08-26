package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	githubRegexp *regexp.Regexp = regexp.MustCompile(`^github\.com`)
	verbose      bool           = true
	gopath       string
)

const VERSION string = "0.1.0"

type Pkg struct {
	Name string
	SHA  string
}

// Grab environment variable(s)
func init() {
	gopath = os.Getenv("GOPATH")
	if gopath == "" {
		Fatal("GOPATH not set")
	}
}

// Print an error and fail
func Fatal(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] ")
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(2)
}

// Get the directory in the GOPATH of the package.
func (pkg Pkg) GitDir() (string, error) {
	path := strings.Split(pkg.Name, "/")
	if len(path) < 3 {
		return "", fmt.Errorf("Bad repo format")
	}
	basepath := filepath.Join(path[:3]...)
	return filepath.Join(gopath, "src", basepath, ".git"), nil
}

// Build a package.
func (pkg Pkg) Build() error {
	goBuild := exec.Command("go", "build", "-v", pkg.Name)
	goBuild.Stderr = os.Stderr
	if verbose {
		goBuild.Stdout = os.Stdout
	}
	return goBuild.Run()
}

// Go get a package.
func (pkg Pkg) Get() error {
	goGet := exec.Command("go", "get", "-v", "-d", pkg.Name)
	goGet.Stderr = os.Stderr
	if verbose {
		goGet.Stdout = os.Stdout
	}
	return goGet.Run()
}

// Get the current SHA of a package.
func (pkg Pkg) CurrSha() (string, error) {
	gitDir, err := pkg.GitDir()
	if err != nil {
		return "", err
	}
	getSha := exec.Command("git", "--git-dir", gitDir, "rev-parse", "HEAD")
	sha, err := getSha.Output()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(sha), " \n"), nil
}

// Revert a package to a specific SHA.
func (pkg Pkg) Revert(sha string) error {
	gitDir, err := pkg.GitDir()
	if err != nil {
		return err
	}
	revert := exec.Command("git", "--git-dir", gitDir, "reset", sha)
	revert.Stderr = os.Stderr
	err = revert.Run()
	if err != nil {
		return fmt.Errorf("Could not reset '%s' to commit '%s'",
			pkg.Name, sha)
	}
	return nil
}

// Read a goopfile and get a list of packages
func ReadGoopfile(Goopfile string) ([]Pkg, error) {
	pkgs := []Pkg{}
	goop, err := ioutil.ReadFile(Goopfile)
	if err != nil {
		return pkgs, err
	}
	lines := strings.Split(strings.Trim(string(goop), "\n"), "\n")
	for i, line := range lines {
		line := strings.Replace(line, "\t", " ", -1)
		pkgInfo := strings.Split(line, " ")
		switch len(pkgInfo) {
		case 0:
			continue
		case 1:
			pkgs = append(pkgs, Pkg{pkgInfo[0], ""})
		case 2:
			pkgs = append(pkgs, Pkg{
				Name: pkgInfo[0],
				SHA:  strings.TrimLeft(pkgInfo[1], "#"),
			})
		default:
			return pkgs, fmt.Errorf(
				"Line %d: invalid number of columns", i)
		}
	}
	return pkgs, nil
}

// Get all github dependencies of a package.
func GetDeps(goFiles []string) ([]Pkg, error) {
	pkgs := []Pkg{}
	args := []string{"list", "-f", "{{.Deps}}"}
	args = append(args, goFiles...)
	listDeps := exec.Command("go", args...)
	listDeps.Stderr = os.Stderr
	depsOut, err := listDeps.Output()
	if err != nil {
		return pkgs, err
	}
	deps := strings.Split(strings.Trim(string(depsOut), "[] \n"), " ")
	for _, dep := range deps {
		if !githubRegexp.MatchString(dep) {
			continue
		}
		pkg := Pkg{dep, ""}
		sha, err := pkg.CurrSha()
		if err != nil {
			return pkgs, fmt.Errorf(
				"Could not get SHA of package '%s'", pkg.Name)
		}
		pkg.SHA = sha
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

// Install subcommand.
func Install(args []string) {
	goopfile := "Goopfile"
	if len(args) > 0 {
		goopfile = args[0]
	}
	pkgs, err := ReadGoopfile(goopfile)
	if err != nil {
		Fatal(err.Error())
	}
	for _, pkg := range pkgs {
		pkg.Get()
	}
	for _, pkg := range pkgs {
		if pkg.SHA == "" {
			continue
		}
		currSha, err := pkg.CurrSha()
		if err != nil {
			Fatal(err.Error())
		}
		if pkg.SHA == currSha {
			continue
		}
		err = pkg.Revert(pkg.SHA)
		if err != nil {
			Fatal(err.Error())
		}
		fmt.Printf("%s (revert: %s)\n", pkg.Name, pkg.SHA)
	}
}

// Freeze subcommand.
func Freeze(args []string) {
	goFiles := args
	pkgs, err := GetDeps(goFiles)
	if err != nil {
		Fatal(err.Error())
	}
	maxLen := 0
	for _, pkg := range pkgs {
		if len(pkg.Name) > maxLen {
			maxLen = len(pkg.Name)
		}
	}
	f := fmt.Sprintf("%%-%ds %%s\n", maxLen)
	for _, pkg := range pkgs {
		fmt.Printf(f, pkg.Name, pkg.SHA)
	}
}

// Print help for gooper.
func PrintHelp() {
	fmt.Println(`gooper your Goopfiles:

commands:

    freeze 
    install
`)
	os.Exit(1)
}

func main() {
	if len(os.Args) == 1 {
		PrintHelp()
	}
	switch os.Args[1] {
	case "install":
		Install(os.Args[2:])
	case "freeze":
		Freeze(os.Args[2:])
	case "help":
		PrintHelp()
	default:
		fmt.Printf("gooper: unknown subcommand '%s'\n", os.Args[1])
		fmt.Printf("Run 'gooper help' for usage.\n")
		os.Exit(1)
	}
}
