package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aereal/lasc"
)

func NewApp() *App {
	return &App{
		outStream: os.Stdout,
		errStream: os.Stderr,
	}
}

type App struct {
	outStream io.Writer
	errStream io.Writer
}

const (
	statusOK int = iota
	statusNG
)

func (a *App) Run(argv []string) int {
	fs := flag.NewFlagSet(filepath.Base(argv[0]), flag.ContinueOnError)
	var opts lasc.Options
	fs.StringVar(&opts.RootDirectory, "root", ".", "root directory")
	switch err := fs.Parse(argv[1:]); err {
	case flag.ErrHelp:
		return statusOK
	case nil:
		// skip
	default:
		return statusNG
	}
	app := lasc.NewApp(opts)
	if err := app.Run(); err != nil {
		fmt.Fprintln(a.errStream, err)
		return statusNG
	}
	return statusOK
}
