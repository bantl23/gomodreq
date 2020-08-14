package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v2"
)

const (
	gomod = "go.mod"
	goreq = "go.req"
)

var (
	version = "v0.0.0"
	commit  = "dev"
)

// Requirements contains the golang code requirements
type Requirements struct {
	Required map[string]string   `yaml:"required,omitempty"`
	Banned   map[string][]string `yaml:"banned,omitempty"`
}

func getData(filename string) ([]byte, error) {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil, err
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func getReqs() (*Requirements, error) {
	reqs := new(Requirements)

	data, err := getData(goreq)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, reqs)
	if err != nil {
		return nil, err
	}

	return reqs, nil
}

func getMod() (*modfile.File, error) {
	data, err := getData(gomod)
	if err != nil {
		return nil, err
	}
	mod, err := modfile.Parse(gomod, data, nil)
	if err != nil {
		return nil, err
	}

	return mod, nil
}

func main() {
	if len(os.Args) == 2 {
		if os.Args[1] == "-v" || os.Args[1] == "--version" {
			fmt.Printf("%s (%s)\n", version, commit)
			os.Exit(0)
		}
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			fmt.Println("Checks modules against requirements list")
			os.Exit(0)
		}
	}

	reqs, err := getReqs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	mod, err := getMod()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	exitValue := 0

	for i := range mod.Require {
		path := mod.Require[i].Mod.Path
		modvers := mod.Require[i].Mod.Version
		_, ok := reqs.Required[path]
		if ok == true {
			re, err := regexp.Compile(reqs.Required[path])
			if err != nil {
				fmt.Printf("Error compiling regular expression: %s\n", reqs.Required[path])
				os.Exit(1)
			} else {
				if !re.Match([]byte(modvers)) {
					fmt.Printf("Error: package %s version %s does not met requirements [regex is %s]\n", path, modvers, reqs.Required[path])
					exitValue = 1
				}
			}

		}
	}

	for i := range mod.Require {
		path := mod.Require[i].Mod.Path
		modvers := mod.Require[i].Mod.Version
		_, ok := reqs.Banned[path]
		if ok == true {
			for j := range reqs.Banned[path] {
				re, err := regexp.Compile(reqs.Banned[path][j])
				if err != nil {
					fmt.Printf("Error compiling regular expression: %s\n", reqs.Required[path])
					os.Exit(1)
				} else {
					if re.Match([]byte(modvers)) {
						fmt.Printf("Error: package %s version %s is banned [regex is %s]\n", path, modvers, reqs.Banned[path][j])
						exitValue = 1
					}
				}
			}
		}
	}

	if exitValue == 0 {
		fmt.Printf("All module requirements met")
	}

	os.Exit(exitValue)
}
