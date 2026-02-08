package main

import (
	"os"

	"github.com/fwartner/prjct/cmd"
)

func main() {
	code := cmd.Execute()
	os.Exit(code)
}
