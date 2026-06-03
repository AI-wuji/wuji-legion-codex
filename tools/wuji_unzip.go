package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: unzip -Z1 <zip> | unzip -p <zip> <entry>")
}

func openArchive(path string) (*zip.ReadCloser, error) {
	resolved, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return zip.OpenReader(resolved)
}

func listEntries(archive *zip.ReadCloser) error {
	for _, entry := range archive.File {
		fmt.Fprintln(os.Stdout, entry.Name)
	}
	return nil
}

func printEntry(archive *zip.ReadCloser, entryName string) error {
	normalized := strings.ReplaceAll(entryName, "\\", "/")
	for _, entry := range archive.File {
		if entry.Name != normalized {
			continue
		}
		handle, err := entry.Open()
		if err != nil {
			return err
		}
		defer handle.Close()
		_, err = io.Copy(os.Stdout, handle)
		return err
	}
	return fmt.Errorf("zip entry not found: %s", normalized)
}

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(1)
	}

	mode := os.Args[1]
	archive, err := openArchive(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer archive.Close()

	switch mode {
	case "-Z1":
		if err := listEntries(archive); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "-p":
		if len(os.Args) < 4 {
			usage()
			os.Exit(1)
		}
		if err := printEntry(archive, os.Args[3]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(1)
	}
}
