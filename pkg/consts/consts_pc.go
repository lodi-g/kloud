// +build linux,amd64

package consts

// Constants used by Kloud throughout the program
const (
	SDMountPoint = "./_sd"
	SyncFolder   = SDMountPoint + "/KloudSync"
	InternalDir  = SDMountPoint + "/.kloud"
	TLSFilePath  = InternalDir + "/cacert.pem"
)
