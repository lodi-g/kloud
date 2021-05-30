#!/bin/sh

internal_dir="%s"
sync_dir="%s"
sd_sync_dir="/mnt/sd/KloudSync"

log_file="$internal_dir/launcher.log"

function log() {
  echo "[$(date)] $1" >> "$log_file"
}

log "Launcher started. internal_dir=$internal_dir, sync_dir=$sync_dir, sd_sync_dir=$sd_sync_dir"

# Make sure we have internet
ping -c2 1.1.1.1 > /dev/null

# Run kloud
log "Starting kloud"
"$internal_dir/kloud"
exit_code=$?
log "kloud exited with code $exit_code"

if [ $exit_code != 0 ]; then
  log "kloud exited with an error, not performing library refresh"
  exit 0
fi

# Trick to refresh library: https://github.com/wernerb/kobo-wget-sync/blob/master/src/usr/local/wget-sync/refresh_library.sh
mkdir -p "$sd_sync_dir"
mount "$kloud_library_folder" "$sd_sync_dir"

log "Refreshing library"
echo sd add /dev/mmcblk1p1 >> /tmp/nickel-hardware-status
log "Done"
