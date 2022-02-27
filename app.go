package lasc

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
)

type Options struct {
	RootDirectory string
}

func NewApp(opts Options) *App {
	return &App{opts: opts}
}

type App struct {
	opts Options
}

func (a *App) Run() error {
	if err := a.initModule(); err != nil {
		return fmt.Errorf("failed to initialize module: %w", err)
	}
	if err := a.buildFiles(); err != nil {
		return fmt.Errorf("failed to build files: %w", err)
	}
	if err := a.formatFiles(); err != nil {
		return fmt.Errorf("failed to format files: %w", err)
	}
	if err := a.installDependencies(); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}
	if err := a.writeFunctionConfig(); err != nil {
		return fmt.Errorf("failed to write function config: %w", err)
	}
	return nil
}

func (a *App) installDependencies() error {
	if err := runCommand(a.opts.RootDirectory, "go", "mod", "tidy"); err != nil {
		return err
	}
	if err := runCommand(a.opts.RootDirectory, "go", "mod", "download"); err != nil {
		return err
	}
	return nil
}

func (a *App) formatFiles() error {
	return runCommand(a.opts.RootDirectory, "go", "fmt")
}

func (a *App) buildFiles() error {
	t, err := template.ParseFS(templates, "**/*.tmpl")
	if err != nil {
		return fmt.Errorf("cannot parse templates: %w", err)
	}
	targets := map[string]map[string]interface{}{
		// template file -> args
		"Dockerfile.tmpl": {
			"BuildDirectory": a.opts.RootDirectory,
		},
		"main.go.tmpl": {},
	}
	for src, args := range targets {
		dest := filepath.Join(a.opts.RootDirectory, src[0:len(src)-5])
		outFile, err := openFileForWrite(dest)
		if err != nil {
			return fmt.Errorf("cannot open file %s: %w", dest, err)
		}
		if err := t.ExecuteTemplate(outFile, src, args); err != nil {
			return fmt.Errorf("failed to execute template (%s): %w", src, err)
		}
	}
	return nil
}

func (a *App) initModule() error {
	if err := os.MkdirAll(a.opts.RootDirectory, 0755); err != nil {
		return fmt.Errorf("cannot create root directory: %w", err)
	}

	if isExist(filepath.Join(a.opts.RootDirectory, "go.mod")) {
		return nil
	}

	if err := runCommand(a.opts.RootDirectory, "go", "mod", "init"); err != nil {
		return fmt.Errorf("failed to run go mod init: %w", err)
	}
	return nil
}

var (
	//go:embed templates
	templates embed.FS
)

func runCommand(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.Run()
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func openFileForWrite(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
}

type codeImage struct {
	ImageUri string
}

type createFunctionInput struct {
	FunctionName string
	PackageType  string
	Role         string
	MemorySize   int
	Publish      bool
	Timeout      int
	Code         codeImage
}

func (a *App) writeFunctionConfig() error {
	dest := filepath.Join(a.opts.RootDirectory, "config.cue")
	if isExist(dest) {
		return nil
	}

	out, err := openFileForWrite(dest)
	if err != nil {
		return err
	}
	if err := createFunctionConfig(out); err != nil {
		return err
	}
	return nil
}

func createFunctionConfig(out io.Writer) error {
	cfi := createFunctionInput{}
	cv, err := fillInput(cfi)
	if err != nil {
		return err
	}
	doc, err := format.Node(cv.Syntax(), format.Simplify(), format.UseSpaces(2))
	if err != nil {
		return fmt.Errorf("failed to format CUE node: %w", err)
	}
	fmt.Fprintln(out, string(doc))
	return nil
}

func fillInput(cfi createFunctionInput) (cue.Value, error) {
	cc := cuecontext.New()
	cv := cc.EncodeType(cfi)
	cv = cv.FillPath(cue.ParsePath("PackageType"), "Image")
	cv = cv.FillPath(cue.ParsePath("MemorySize"), 128)
	cv = cv.FillPath(cue.ParsePath("Timeout"), 10)
	cv = cv.FillPath(cue.ParsePath("Publish"), true)
	return cv, nil
}
