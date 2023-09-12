package main

import (
	"github.com/achetronic/autoheater/internal/cmd"
	"os"
	"path/filepath"
)

func main() {
	baseName := filepath.Base(os.Args[0])

	err := cmd.NewAutoheaterCommand(baseName).Execute()
	cmd.CheckError(err)
}
