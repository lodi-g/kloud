package main

import (
	_ "embed"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"kloud/pkg/config"
	"kloud/pkg/consts"
	"kloud/pkg/nextcloud"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

//go:embed cacert.pem
var cacert []byte

func getLocalFiles(root string) (map[string]int64, error) {
	ret := map[string]int64{}

	// Walk the local filesystems and return a map[filename]filesize
	err := filepath.Walk(root, func(path string, fileinfo fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fileinfo.IsDir() {
			return nil
		}

		relativePath := strings.ReplaceAll(path, consts.SyncFolder+"/", "")
		ret[relativePath] = fileinfo.Size()
		return nil
	})

	if err != nil {
		return nil, err
	}
	return ret, nil
}

func diffFiles(local, remote map[string]int64) (toDownload, toDelete []string) {
	// Find what files should be downloaded from the remote server
	for remoteFileName, remoteFileSize := range remote {
		localFileSize, localFileExists := local[remoteFileName]
		if localFileExists == false || localFileSize != remoteFileSize {
			toDownload = append(toDownload, remoteFileName)
		}
	}

	// Find what files should be deleted from the local filesystem
	for localFileName := range local {
		_, remoteFileExists := remote[localFileName]
		if remoteFileExists == false {
			toDelete = append(toDelete, localFileName)
		}
	}

	return toDownload, toDelete
}

func downloadFiles(client nextcloud.Client, files []string) error {
	// Iterate over the files and download each one into the sync directory
	for _, fileName := range files {
		// Download file
		fileContent, err := client.DownloadFile(fileName)
		if err != nil {
			return err
		}

		// Create directory if needed
		fullPath := filepath.Join(consts.SyncFolder, fileName)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}

		// Write file
		if err := ioutil.WriteFile(fullPath, fileContent, 0600); err != nil {
			return err
		}
	}

	return nil
}

func deleteFiles(files []string) error {
	// Iterate over the list of files and delete them
	for _, fileName := range files {
		// Delete the file
		fullPath := filepath.Join(consts.SyncFolder, fileName)
		if err := os.Remove(fullPath); err != nil {
			return err
		}

		// Delete the directory if it's empty (and not the root directory)
		dir := filepath.Dir(fullPath)
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return err
		}

		if len(files) == 0 && dir != consts.SyncFolder {
			os.Remove(dir)
		}
	}

	return nil
}

func init() {
	logFilePath := consts.InternalDir + "/" + "kloud.log"

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	logger.Out = file
}

func main() {
	// Start and read config
	config, err := config.Get()
	if err != nil {
		logger.WithField("error", err).Fatal("Cannot retrieve configuration")
		os.Exit(1)
	}
	logger.Infof("Started with configuration: %+v", config)

	// Get the list of files in the sync directory
	localFiles, err := getLocalFiles(consts.SyncFolder)
	if err != nil {
		logger.WithField("error", err).Fatal("Cannot read local filesystem")
		os.Exit(1)
	}
	logger.WithField("local_files", localFiles).Info("Retrieved local files")

	// Create the nextcloud client and use it to get the list of files in the server
	ncClient, err := nextcloud.NewClient(cacert, config.Server, config.ShareID)
	if err != nil {
		logger.WithField("error", err).Fatal("Impossible to create NextCloud client")
		os.Exit(1)
	}

	remoteFiles, err := ncClient.GetRemoteFiles()
	if err != nil {
		logger.WithField("error", err).Fatal("Cannot load remote NextCloud")
		os.Exit(1)
	}
	logger.WithField("remote_files", remoteFiles).Info("Retrieved remote files")

	// Compute the files to download and to delete, and download and deletes them
	toDownload, toDelete := diffFiles(localFiles, remoteFiles)
	logger.WithField("to_download", toDownload).Info("Files to download")
	logger.WithField("to_delete", toDelete).Info("Files to delete")

	if err := downloadFiles(ncClient, toDownload); err != nil {
		logger.WithField("error", err).Fatal("Failed to download files")
	}
	if err := deleteFiles(toDelete); err != nil {
		logger.WithField("error", err).Fatal("Failed to delete files")
	}

	logger.Info("Success")
}
