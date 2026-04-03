package main

import (
	"os"

	"github.com/spaulg/solo/cmd/solo/subcommand"
)

func main() {
	os.Exit(subcommand.Execute())
}
