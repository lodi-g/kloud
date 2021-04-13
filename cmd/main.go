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

// TODO: Synchronize deletion?

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

func diffFiles(local, remote map[string]int64) (ret []string) {
	for remoteFileName, remoteFileSize := range remote {
		localFileSize, localFileExists := local[remoteFileName]
		if localFileExists == false || localFileSize != remoteFileSize {
			ret = append(ret, remoteFileName)
		}
	}

	return ret
}

func syncFiles(client nextcloud.Client, files []string) error {
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

	toSyncFiles := diffFiles(localFiles, remoteFiles)
	logger.WithField("to_sync_files", toSyncFiles).Info("Files to sync")

	if err := syncFiles(ncClient, toSyncFiles); err != nil {
		logger.WithField("error", err).Fatal("Failed to synchronize files")
	}
	logger.Info("Success")
}
