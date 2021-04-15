package main

import (
	"archive/tar"
	"compress/gzip"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	configTpl = `server: %s
share: %s
`
	archiveFilename = "KoboRoot.tgz"
)

//go:embed kloud
var kloudBinary []byte

//go:embed 97-kloud.rules
var udevRules []byte

func help() {
	fmt.Printf("Kloud bootstraper\n\n")

	fmt.Println("This program generates a KoboRoot.tgz archive which you have to")
	fmt.Println("copy to your .kobo/ folder of your Kobo device. This archive")
	fmt.Printf("will install kloud to your device.\n\n")

	fmt.Printf("Usage: %s ServerURL ShareID\n", os.Args[0])
	fmt.Printf("Example: %s https://cloud.server.com eagB90Oy5uUa4eB\n\n", os.Args[0])

	fmt.Println("The share ID is obtained from the share URL in NextCloud.")
	fmt.Println("Example: https://cloud.server.com/s/eagB90Oy5uUa4eB.")
}

func prepareFolderToArchive(wd, config string) {
	mntKloudPath := path.Join(wd, "mnt", "onboard", ".kloud")
	if err := os.MkdirAll(mntKloudPath, os.ModePerm); err != nil {
		panic(err)
	}

	if err := os.WriteFile(path.Join(mntKloudPath, "kloud"), kloudBinary, 0777); err != nil {
		panic(err)
	}

	if err := os.WriteFile(path.Join(mntKloudPath, "config.yml"), []byte(config), 0644); err != nil {
		panic(err)
	}

	kloudSyncPath := path.Join(wd, "mnt", "onboard", "KloudSync")
	if err := os.MkdirAll(kloudSyncPath, os.ModePerm); err != nil {
		panic(err)
	}

	udevPath := path.Join(wd, "etc", "udev", "rules.d")
	if err := os.MkdirAll(udevPath, os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.WriteFile(path.Join(udevPath, "97-kloud.rules"), udevRules, 0644); err != nil {
		panic(err)
	}
}

func createArchive(wd string) {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	target := path.Join(cwd, archiveFilename)

	f, err := os.Create(target)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()
	tarw := tar.NewWriter(gzw)
	defer tarw.Close()

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}

		if path == wd {
			return nil
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			panic(err)
		}
		header.Name = strings.ReplaceAll(path, wd+"/", "")

		if err := tarw.WriteHeader(header); err != nil {
			panic(err)
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		if _, err := io.Copy(tarw, file); err != nil {
			panic(err)
		}

		return nil
	}

	filepath.Walk(wd, walkFunc)
}

func main() {
	if len(os.Args) != 3 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		help()
		os.Exit(0)
	}

	wd, err := os.MkdirTemp("", "kloud-bootstraper")
	if err != nil {
		log.Fatalf("Unable to create working directory: %s", err)
		os.Exit(1)
	}
	defer os.RemoveAll(wd)

	config := fmt.Sprintf(configTpl, os.Args[1], os.Args[2])
	prepareFolderToArchive(wd, config)
	createArchive(wd)

	fmt.Printf("A %s file was created, copy it to your .kobo folder to apply the update\n", archiveFilename)
}
