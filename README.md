# kloud

## Description

Kloud is a tool used to synchronize a Kobo e-reader with a remote NextCloud server.

When you connect to Wi-Fi, it will scan the local filesystem and the remote NextCloud server and compute what is different between the two. It will download new files (or files that have the same name but not the same size), and delete files that were deleted from the remote server.

## Things you should be aware of

- This has been tested against a Kobo Clara HD. It *should not* break other Kobo devices, but it may not work. If it does not, please file a issue and specify what device you are using.

- There is no uninstall procedure right now. You can delete yourself the directories and files created by the `KoboRoot.tgz` file. You need `telnet` or `ftp` access to delete the udev rules.

## Installation

### Preparation

Run the bootstrap program (included in the release) like so, where serverURL is your NextCloud's server URL (e.g. https://cloud.example.com) and shareID is the end of the share link (e.g. https://cloud.example.com/s/abcd1234 - shareID is abcd1234).

```
./bootstrap serverURL shareID
```

This will generate a `KoboRoot.tgz` archive.

### Installation

Put the `KoboRoot.tgz` file in your Kobo's `.kobo` directory, and reboot your device.
