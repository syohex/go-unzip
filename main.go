package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"code.google.com/p/mahonia"
)

func replaceWindowsPathSeparator(path string) string {
	bs := regexp.MustCompile(`\\`)
	unix := bs.ReplaceAllLiteralString(path, "/")

	driveRe := regexp.MustCompile(`\A[A-Z]/`)
	return driveRe.ReplaceAllLiteralString(unix, "")
}

func isDirectoryExisted(dir string) bool {
	f, err := os.Open(dir)
	if err != nil {
		return false
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return false
	}

	return fi.Mode().IsDir()
}

func convertToUtf8(path string, encoding string) string {
	return mahonia.NewDecoder(encoding).ConvertString(path)
}

func unzip(src string, encoding string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	fmt.Printf("Extract %s\n", src)
	for _, zf := range r.File {
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := replaceWindowsPathSeparator(zf.Name)
		name := convertToUtf8(path, encoding)

		parent := filepath.Dir(name)

		fmt.Printf("    %s\n", name)
		if !isDirectoryExisted(parent) {
			os.MkdirAll(parent, 0755)
		}

		f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zf.Mode())
		if err != nil {
			return err
		}

		_, err = io.Copy(f, rc)
		if err != nil {
			return err
		}

		f.Close()
	}

	return nil
}

func main() {
	var encoding string
	flag.StringVar(&encoding, "encoding", "cp932", "file name encoding(default: cp932)")

	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: go-unzip file\n")
		os.Exit(1)
	}

	file := args[0]
	err := unzip(file, encoding)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed: unzip '%s'\n", file)
		log.Fatal(err)
	}
}
