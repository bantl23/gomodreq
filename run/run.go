package run

import (
	"fmt"
	"html"
	"regexp"
	"strconv"

	"github.com/bantl23/gomodreq/modinfo"
	"github.com/bantl23/gomodreq/reqinfo"
)

var (
	check = html.UnescapeString("\033[32m" + "&#" + strconv.Itoa(0x2705) + ";" + "\033[0m")
	cross = html.UnescapeString("\033[31m" + "&#" + strconv.Itoa(0x274e) + ";" + "\033[0m")
)

func checkRequired(ri *reqinfo.ReqInfo, mi []*modinfo.ModulePublic, exitCode int) (int, error) {
	for i := range mi {
		path := mi[i].Path
		modvers := mi[i].Version
		_, ok := ri.Required[path]
		if ok == true {
			icon := cross
			if ri.Required[path] == "latest" {
				if mi[i].Update != nil {
					exitCode = 1
				}
			} else {
				re, err := regexp.Compile(ri.Required[path])
				if err != nil {
					return exitCode, fmt.Errorf("unable to compile regex %s [%+v]", ri.Required[path], err)
				}
				if !re.Match([]byte(modvers)) {
					exitCode = 1
				} else {
					icon = check
				}
			}
			fmt.Printf("required: %s %s\n", path, icon)
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
				icon := cross
				if ri.Banned[path][j] == "latest" && mi[i].Update == nil {
					exitCode = 2
				} else {
					re, err := regexp.Compile(ri.Banned[path][j])
					if err != nil {
						return exitCode, fmt.Errorf("unable to compile regex %s [%+v]", ri.Banned[path][j], err)
					}
					if re.Match([]byte(modvers)) {
						exitCode = 2
					} else {
						icon = check
					}
				}
				fmt.Printf("banned:   %s %s\n", path, icon)
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

	icon := cross
	if reqExitCode == 0 && banExitCode == 0 {
		icon = check
	}
	fmt.Printf("all:      %s\n", icon)

	return reqExitCode + banExitCode
}
