package run

import (
	"fmt"
	"regexp"

	"github.com/bantl23/gomodreq/modinfo"
	"github.com/bantl23/gomodreq/reqinfo"
)

func checkRequired(ri *reqinfo.ReqInfo, mi []*modinfo.ModulePublic, exitCode int) (int, error) {
	for i := range mi {
		path := mi[i].Path
		modvers := mi[i].Version
		_, ok := ri.Required[path]
		if ok == true {
			if ri.Required[path] == "latest" {
				if mi[i].Update != nil {
					fmt.Printf("error package %s version %s is not the latest version [latest=%s]\n", path, modvers, mi[i].Update.Version)
					exitCode = 1
				}
			} else {
				re, err := regexp.Compile(ri.Required[path])
				if err != nil {
					return exitCode, fmt.Errorf("unable to compile regex %s [%+v]", ri.Required[path], err)
				}
				if !re.Match([]byte(modvers)) {
					fmt.Printf("error package %s version %s does not met requirements [regex is %s]\n", path, modvers, ri.Required[path])
					exitCode = 1
				} else {
					fmt.Printf("required: %s met\n", path)
				}
			}
		}
	}
	return exitCode, nil
}

func checkBanned(ri *reqinfo.ReqInfo, mi []*modinfo.ModulePublic, exitCode int) (int, error) {
	for i := range mi {
		path := mi[i].Path
		modvers := mi[i].Version
		_, ok := ri.Banned[path]
		if ok == true {
			for j := range ri.Banned[path] {
				if ri.Banned[path][j] == "latest" && mi[i].Update == nil {
					fmt.Printf("error package %s version %s is the latest version which is banned\n", path, modvers)
					exitCode = 2
				} else {
					re, err := regexp.Compile(ri.Banned[path][j])
					if err != nil {
						return exitCode, fmt.Errorf("unable to compile regex %s [%+v]", ri.Banned[path][j], err)
					}
					if re.Match([]byte(modvers)) {
						fmt.Printf("error package %s version %s is banned [regex is %s]\n", path, modvers, ri.Banned[path][j])
						exitCode = 2
					} else {
						fmt.Printf("banned: %s met\n", path)
					}
				}
			}
		}
	}
	return exitCode, nil
}

// Run runs the checks
func Run(reqLoc []string) int {
	mod, err := modinfo.GetModInfo()
	if err != nil {
		fmt.Printf("getting modules [%+v]\n", err)
		return -1
	}

	reqExitCode := 0
	banExitCode := 0
	for i := range reqLoc {
		fmt.Printf("checking requirements: %s\n", reqLoc[i])
		req, err := reqinfo.GetReqInfo(reqLoc[i])
		if err != nil {
			fmt.Printf("error getting requirements [%+v]\n", err)
			return -1
		}
		reqExitCode, err = checkRequired(req, mod, reqExitCode)
		if err != nil {
			fmt.Printf("error checking required modules [%+v]\n", err)
			return -1
		}
		banExitCode, err = checkBanned(req, mod, banExitCode)
		if err != nil {
			fmt.Printf("error checking banned modules [%+v]\n", err)
			return -1
		}
	}

	if reqExitCode == 0 && banExitCode == 0 {
		fmt.Println("All module requirements met")
	}

	return reqExitCode + banExitCode
}
