/*
Gov is the GO Version manager.

Usage:
	gov install (version)    # Installs a specific Go version.
	                         # The directories that installed are as follows:
	                         #     Linux:   ~/.goversions/[version]/
	                         #     Windows: %APPDATA%\.goversions\[version]\
	                         # ... or $GOVERSIONS_PATH(%GOVERSIONS_PATH% on Windows).
	gov uninstall (version)  # Uninstall version.
	gov ls                   # List the installed versions.
	gov (version) [command]  # Run the command with specific version of Go binary.
	gov help                 # Print this message
*/
package main

//go:generate go run github.com/ybkimm/gov/utils/helpgen

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const (
	archivePath = "https://dl.google.com/go/go%version%.%os%-%arch%.zip"
)

var regexGoVersion = regexp.MustCompile(`(?:v|go)?([1-9][0-9]*\.[0-9]+\.[0-9]+)`)

var basePath string

func init() {
	env := os.Getenv("GOVERSIONS_PATH")
	if len(env) > 0 {
		p, err := filepath.Abs(env)
		if err != nil {
			panic(err)
		}
		basePath = p
	} else {
		p, err := filepath.Abs(filepath.Join(os.Getenv("APPDATA"), ".goversions"))
		if err != nil {
			panic(err)
		}
		basePath = p
	}
}

func main() {
	args := os.Args
	if len(args) < 2 {
		help()
		os.Exit(1)
	}

	var err error

	switch args[1] {
	case "install":
		err = withVersion(args[2:], install)

	case "uninstall":
		err = withVersion(args[2:], uninstall)

	case "ls":
		err = listVersions()

	case "help":
		err = help()

	default:
		err = withVersion(args[1:], runCommand)
		if errors.Is(err, ErrVersionRequired) || errors.Is(err, ErrInvalidVersion) {
			fmt.Printf("Unknown command: %s\n", args[1])
			err = ErrUnknownCommand
		}
	}

	switch err {
	case nil:
		os.Exit(0)

	case
		ErrUnknownCommand,
		ErrHelped,
		ErrVersionRequired,
		ErrInvalidVersion,
		ErrUnknownArgs,
		os.ErrExist,
		os.ErrNotExist,
		io.EOF:

		os.Exit(1)

	default:
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func install(version string, args []string) error {
	var (
		useDefault = false
	)

	for _, arg := range args {
		switch arg {
		case "--default":
			useDefault = true

		default:
			return ErrUnknownArgs
		}
	}

	// Check dist directory and create if not exists
	var distDir string
	if useDefault {
		distDir = defaultDirectory
	} else {
		distDir = installPath(version)
	}

	fmt.Printf("Target directory: %s\n", distDir)

	_, err := os.Stat(distDir)
	if !os.IsNotExist(err) {
		if useDefault {
			os.RemoveAll(distDir)
		} else {
			fmt.Printf("Directory %s is already exists!\n", distDir)
			return os.ErrExist
		}
	}

	err = os.MkdirAll(distDir, 755)
	if err != nil {
		return err
	}

	// Download archive...
	fmt.Println("Downloading archive...")
	file, err := download(version)
	if err != nil {
		return err
	}
	defer os.Remove(file)

	fmt.Println("Unzipping archive...")
	err = Unzip(file, distDir)
	if err != nil {
		return err
	}

	fmt.Println("Done!")
	return nil
}

func uninstall(version string, _ []string) error {
	p := installPath(version)

	_, err := os.Stat(p)
	if os.IsNotExist(err) {
		fmt.Printf("go%s is not installed!\n", version)
		return os.ErrNotExist
	}

	return os.RemoveAll(p)
}

func listVersions() error {
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No version is installed.")
			return nil
		} else {
			return err
		}
	}

	for _, file := range files {
		caps := regexGoVersion.FindStringSubmatch(file.Name())
		if caps == nil {
			continue
		}
		fmt.Printf("%s\n", caps[1])
	}

	return nil
}

func runCommand(version string, args []string) error {
	cmd := exec.Command(filepath.Join(installPath(version), gobin), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func help() error {
	//noinspection GoPrintFunctions
	fmt.Print(helpMessage)
	return ErrHelped
}

func parseVersion(args []string) (string, error) {
	if len(args) == 0 {
		return "", ErrVersionRequired
	}

	caps := regexGoVersion.FindStringSubmatch(args[0])
	if caps == nil {
		return "", fmt.Errorf("%w: %s", ErrInvalidVersion, args[0])
	}

	return caps[1], nil
}

func withVersion(args []string, fn func(string, []string) error) error {
	version, err := parseVersion(args)
	if err != nil {
		return err
	}

	return fn(version, args[1:])
}

func download(version string) (string, error) {
	uri := strings.NewReplacer(
		"%version%", version,
		"%os%", runtime.GOOS,
		"%arch%", runtime.GOARCH,
	).Replace(archivePath)

	// Create temporary file
	tmpFile, err := ioutil.TempFile("", path.Base(uri))
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	defer tmpFile.Close()

	// Send request
	resp, err := http.Get(uri)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("download: %w", err)
	}

	return tmpFile.Name(), nil
}

func installPath(version string) string {
	return filepath.Join(basePath, version)
}
