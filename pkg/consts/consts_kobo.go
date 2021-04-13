// +build linux,arm

package consts

// Constants used by Kloud throughout the program
const (
	SDMountPoint = "/mnt/onboard"
	SyncDir      = SDMountPoint + "/KloudSync"
	InternalDir  = SDMountPoint + "/.kloud"
	TLSFilePath  = InternalDir + "/cacert.pem"
)
