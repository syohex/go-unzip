package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

func convertToUtf8(path string, encoding string) (string, error) {
	var transformer transform.Transformer

	switch encoding {
	case "utf-8", "utf8":
		return path, nil
	case "cp932":
		transformer = japanese.ShiftJIS.NewDecoder()
	case "euc-jp":
		transformer = japanese.EUCJP.NewDecoder()
	case "iso2022":
		transformer = japanese.ISO2022JP.NewDecoder()
	default:
		return "", fmt.Errorf("unsupported encoding: %s", encoding)
	}

	ret, _, err := transform.String(transformer, path)
	if err != nil {
		return "", err
	}

	return ret, nil
}

func unzip(src string, encoding string, list bool) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	fmt.Printf("Extract %s\n", src)
	for _, zf := range r.File {
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		if zf.FileInfo().IsDir() {
			continue
		}

		path := replaceWindowsPathSeparator(zf.Name)
		name, err := convertToUtf8(path, encoding)
		if err != nil {
			return err
		}

		dest := filepath.Join(cwd, name)
		if !strings.HasPrefix(dest, filepath.Clean(cwd)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", name)
		}

		fmt.Printf("    %s\n", name)
		if list {
			continue
		}

		parent := filepath.Dir(dest)
		if !isDirectoryExisted(parent) {
			err := os.MkdirAll(parent, 0755)
			if err != nil {
				return err
			}
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
	var list bool
	flag.StringVar(&encoding, "encoding", "cp932", "file name encoding(default: cp932)")
	flag.BoolVar(&list, "list", false, "List extract files, not extract")

	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: go-unzip file\n")
		os.Exit(1)
	}

	file := args[0]
	err := unzip(file, encoding, list)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed: unzip '%s'\n", file)
		log.Fatal(err)
	}
}
