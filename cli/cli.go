package cli

import (
	"io"
	"os"
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
	return statusOK
}
