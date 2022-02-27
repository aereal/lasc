package main

import (
	"os"

	"github.com/aereal/lasc/cli"
)

func main() {
	os.Exit(cli.NewApp().Run(os.Args))
}
