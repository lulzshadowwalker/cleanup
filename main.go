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
	flag.StringVar(&path, "path", "", "path to the directory to cleanup (can be set as env.DEV_PATH)")
	flag.IntVar(&depth, "depth", 4, "max recursion depth for the specified directory")
	flag.Usage = func() {
		fmt.Printf("Usage of flutter_cleanup↳ runs `flutter clean` for any flutter projects in the specified path within a certain depth\n\n")
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
	initFs, initFsErr := fileSize(path)
	if initFsErr != nil {
		fmt.Print("🦑 cannot read directory size\n")
	} else {
		fmt.Printf("🥑 initial size on disk %d MB\n\n", initFs)
	}

	scan(path, depth)

	finalFs, finalFsErr := fileSize(path)
	if finalFsErr != nil {
		fmt.Print("\n🦑 cannot read directory size\n")
	} else {
		fmt.Printf("\n🥑 final size on disk %d MB ", initFs)
		if initFsErr == nil {
			fmt.Printf("( -%d MB )\n", finalFs-initFs)
		}
	}
}

func scan(path string, depth int) {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("🧣 cannot read given directory %q\n", err)
		return
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

			fs, errFileSize := fileSize(path)
			if errFileSize != nil {
				fmt.Printf("\t🦑 cannot read file size %q..\n", errFileSize)
			} else {
				fmt.Printf("\t↳ size on disk: %d MB\n", fs)
			}

			fmt.Printf("\t↳ running flutter clean ..\n")

			cmd := exec.Command("flutter", "clean")
			cmd.Dir = path
			err = cmd.Run()
			if err != nil {
				fmt.Printf("🌹 cannot execute `flutter clean` in %s %q\n", path, err)
				return
			}

			fmt.Printf("\t↳ 🐸 %s cleaned up successfully\n", path)
			return
		}

		nextPath := filepath.Join(path, e.Name())
		shouldAvoid := contains(nextPath, "node_modules", "vendor", ".git", ".vscode", ".idea", ".npmignore")
		if depth != 0 && !shouldAvoid {
			scan(nextPath, depth-1)
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

func fileSize(path string) (int64, error) {
	s, err := os.Stat(path)
	if err != nil {
		return -1, fmt.Errorf("cannot read file stats %w", err)
	}

	return s.Size(), nil
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
