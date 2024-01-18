package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	path  string
	depth int
)

func init() {
	flag.StringVar(&path, "path", "", "path to the directory to cleanup")
	flag.IntVar(&depth, "depth", 4, "max recursion depth for the specified directory")
	flag.Usage = func() {
		fmt.Printf("Usage of flutter_cleanup\n↳ runs `flutter clean` for any flutter projects in the specified path within a certain depth\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if path == "" {
		p := os.Getenv("DEV_PATH")
		if p == "" {
			fmt.Println("--path to directory has to be specified")
			os.Exit(1)
		}

		path = p
	}
}

func main() {
	fmt.Printf("🍄 running in %s with a depth of %d\n", path, depth)
	scanDir(path, depth)
}

func scanDir(path string, depth int) {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("cannot read given directory %q", err)
		os.Exit(1)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		pubspecPath := filepath.Join(path, "pubspec.yaml")
		isFlutterProject, err := fileExists(pubspecPath)
		if err != nil {
			fmt.Println("🌹", err)
			continue
		}

		if isFlutterProject {
			fmt.Printf("💡 flutter project found (%s)\n", path)
			fmt.Printf("\t↳ running flutter clean ..\n")

			cmd := exec.Command("flutter", "clean")
			cmd.Dir = path
			err = cmd.Run()
			if err != nil {
				fmt.Printf("🌹 cannot execute `flutter clean` in %s %q\n", path, err)
			}

			fmt.Printf("\t↳ 🐸 %s cleaned up successfully\n", path)
			return
		}

		nextPath := filepath.Join(path, e.Name())
		if depth != 0 && !contains("node_modules", "vendor", ".git", ".vscode", ".idea") {
			scanDir(nextPath, depth-1)
		}
	}
}

func contains(haystack string, needle ...string) bool {
	for _, n := range needle {
		if strings.Contains(haystack, n) {
			return true
		}
	}

	return false
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, fmt.Errorf("cannot read file stats %w", err)
	}

	return true, nil
}
