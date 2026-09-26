// Command coj is the CLI for CandyOJ, an online judge for teaching.
//
// It mirrors gh and follows a REST resource model:
//
//	coj <resource> <action> [flags]
//
// For example:
//
//	coj auth login --username alice
//	coj match list --page 2
//	coj match rank --field F_MatchID=42 --field F_ClassID=7
//	coj subject list --json
//	coj api POST /Subject/getSubjectListNew --field page=1
//
// All 190 backend endpoints are reachable: resource commands cover them all,
// and coj api serves as a general escape hatch.
package main

import (
	"fmt"
	"os"

	"github.com/HFunction2013/coj-cli/internal/cli"
	"github.com/HFunction2013/coj-cli/internal/cmd"
)

func main() {
	router := cmd.New()
	if err := router.Execute(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "coj: %v\n", err)
		os.Exit(1)
	}
}

var _ = cli.Command{}
