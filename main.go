package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bantl23/gomodreq/run"
)

// set by build flags
var (
	version = "v0.0.0"
	commit  = "dev"
)

func mainWithExit() int {
	path, err := os.Getwd()
	if err != nil {
		fmt.Println("Fatal: unable to get current directory")
		return -1
	}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage:\n%s <modreq_uri> ... (default: file://%s/.gomodreq.yml):\n",
			filepath.Base(os.Args[0]),
			path,
		)
		flag.PrintDefaults()
	}

	versFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versFlag == true {
		fmt.Printf("%s (%s)\n", version, commit)
		return 0
	}

	reqLoc := os.Args[1:]
	if len(reqLoc) == 0 {
		reqLoc = []string{"file://" + path + "/.gomodreq.yml"}
	}

	return run.Run(reqLoc)
}

func main() {
	os.Exit(mainWithExit())
}
