package main

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"kloud/pkg/config"
	"kloud/pkg/consts"
	"kloud/pkg/nextcloud"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

/*
Installation:

KoboRoot with udev rules

.kloud/kloud
.kloud/kloud.log
.kloud/config.yml
*/

func getLocalFiles(root string) (map[string]int64, error) {
	ret := map[string]int64{}

	walkFunc := func(path string, fileinfo fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fileinfo.IsDir() {
			return nil
		}

		prefixLen := len(consts.SyncFolder) - 1
		ret[path[prefixLen:]] = fileinfo.Size()
		return nil
	}

	if err := filepath.Walk(root, walkFunc); err != nil {
		return nil, err
	}

	return ret, nil
}

func diffFiles(local, remote map[string]int64) (toDownload, toDelete []string) {
	for remoteFileName, remoteFileSize := range remote {
		localFileSize, localFileExists := local[remoteFileName]
		if localFileExists == false || localFileSize != remoteFileSize {
			toDownload = append(toDownload, remoteFileName)
		}
	}

	for localFileName := range local {
		_, remoteFileExists := remote[localFileName]
		if remoteFileExists == false {
			toDelete = append(toDelete, localFileName)
		}
	}

	return toDownload, toDelete
}

func downloadFiles(client nextcloud.Client, files []string) error {
	for _, fileName := range files {
		fileContent, err := client.DownloadFile(fileName)
		if err != nil {
			return err
		}

		fullPath := path.Join(consts.SyncFolder, fileName)
		dir := path.Dir(fullPath)

		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}

		if err := ioutil.WriteFile(fullPath, fileContent, 0644); err != nil {
			return err
		}
	}

	return nil
}

func deleteFiles(files []string) error {
	for _, fileName := range files {
		fullPath := path.Join(consts.SyncFolder, fileName)

		if err := os.Remove(fullPath); err != nil {
			return err
		}

		dir := path.Dir(fullPath)
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return err
		}

		if len(files) == 0 {
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
	config, err := config.Get()
	if err != nil {
		logger.WithField("error", err).Fatal("Cannot retrieve configuration")
		os.Exit(1)
	}

	logger.Infof("Started with configuration: %+v", config)

	localFiles, err := getLocalFiles(consts.SyncFolder)
	if err != nil {
		logger.WithField("error", err).Fatal("Cannot read local filesystem")
		os.Exit(1)
	}
	logger.WithField("local_files", localFiles).Info("Retrieved local files")

	ncClient, err := nextcloud.NewClient(consts.TLSFilePath, config.Server, config.ShareID)
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
