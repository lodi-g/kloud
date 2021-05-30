package main

import (
	"archive/tar"
	"compress/gzip"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"kloud/pkg/consts"
)

const (
	configTpl = `server: %s
share: %s
`
	archiveFilename = "KoboRoot.tgz"
)

//go:embed kloud
var kloudBinary []byte

//go:embed launcher.tpl.sh
var launcherScriptTpl string

//go:embed 97-kloud.tpl.rules
var udevRulesTpl string

func help() {
	fmt.Printf("Usage: %s ServerURL ShareID\n", os.Args[0])
	fmt.Printf("Example: %s https://cloud.server.com eagB90Oy5uUa4eB\n\n", os.Args[0])

	fmt.Println("The share ID is obtained from the share URL in NextCloud.")
	fmt.Println("Example: https://cloud.server.com/s/eagB90Oy5uUa4eB.")
}

func prepareFolderToArchive(wd, config string) {
	// Create .kloud
	mntKloudPath := path.Join(wd, consts.SDMountPoint)
	if err := os.MkdirAll(mntKloudPath, os.ModePerm); err != nil {
		log.Fatalf("Error creating .kloud directory: %v\n", err)
	}

	// Copy binary, generated config to .kloud
	if err := os.WriteFile(path.Join(mntKloudPath, "kloud"), kloudBinary, 0777); err != nil {
		log.Fatalf("Error writing Kloud binary: %v\n", err)
	}
	if err := os.WriteFile(path.Join(mntKloudPath, "config.yml"), []byte(config), 0644); err != nil {
		log.Fatalf("Error writing Kloud config: %v\n", err)
	}
	// Generate launcher script and copy to .kloud
	launcherScript := fmt.Sprintf(launcherScriptTpl, consts.InternalDir, consts.SyncDir)
	if err := os.WriteFile(path.Join(mntKloudPath, "launcher.sh"), []byte(launcherScript), 0644); err != nil {
		log.Fatalf("Error writing launcher script: %v\n", err)
	}

	// Create KloudSync (empty)
	kloudSyncPath := consts.SyncDir
	if err := os.MkdirAll(kloudSyncPath, os.ModePerm); err != nil {
		log.Fatalf("Error creating KloudSync directory: %v\n", err)
	}

	// Create udev rules directory
	udevPath := path.Join(wd, "etc", "udev", "rules.d")
	if err := os.MkdirAll(udevPath, os.ModePerm); err != nil {
		log.Fatalf("Error creating udev directory: %v\n", err)
	}

	// Format and copy to udev rules dir
	launcherScriptPath := path.Join(consts.SDMountPoint, "launcher.sh")
	udevRules := fmt.Sprintf(udevRulesTpl, launcherScriptPath, launcherScriptPath)
	if err := os.WriteFile(path.Join(udevPath, "97-kloud.rules"), []byte(udevRules), 0644); err != nil {
		log.Fatalf("Error writing udev rules: %v\n", err)
	}
}

func createArchive(wd string) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error retrieving cwd: %v\n", err)
	}
	target := path.Join(cwd, archiveFilename)

	f, err := os.Create(target)
	if err != nil {
		log.Fatalf("Error creating archive file: %v\n", err)
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()
	tarw := tar.NewWriter(gzw)
	defer tarw.Close()

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("Error scanning file %s: %v\n", path, err)
		}

		if path == wd {
			return nil
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			log.Fatalf("Error generating tar header for file %s: %v\n", path, err)
		}
		header.Name = strings.ReplaceAll(path, wd+"/", "")

		if err := tarw.WriteHeader(header); err != nil {
			log.Fatalf("Error writing tar header for file %s: %s\n", path, err)
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			log.Fatalf("Error opening file %v: %v\n", path, err)
		}
		defer file.Close()

		if _, err := io.Copy(tarw, file); err != nil {
			log.Fatalf("Error writing %v to tar archive: %v\n", path, err)
		}

		return nil
	}

	filepath.Walk(wd, walkFunc)
}

func main() {
	serverURL := flag.String("server-url", "", "URL of the NextCloud server")
	shareID := flag.String("share-id", "", "Share ID of your NextCloud shared directory")
	flag.Usage = func() {
		fmt.Printf("Kloud bootstraper\n\n")

		fmt.Println("This program generates a KoboRoot.tgz archive which you have to")
		fmt.Println("copy to your .kobo/ folder of your Kobo device. This archive")
		fmt.Printf("will install kloud to your device.\n\n")

		flag.PrintDefaults()
	}

	flag.Parse()

	if *serverURL == "" || *shareID == "" {
		flag.Usage()
		os.Exit(1)
	}

	wd, err := os.MkdirTemp("", "kloud-bootstraper")
	if err != nil {
		log.Fatalf("Unable to create working directory: %s", err)
		os.Exit(1)
	}
	defer os.RemoveAll(wd)

	config := fmt.Sprintf(configTpl, *serverURL, *shareID)
	prepareFolderToArchive(wd, config)
	createArchive(wd)

	fmt.Printf("A %s file was created, copy it to your .kobo folder to apply the update\n", archiveFilename)
}
