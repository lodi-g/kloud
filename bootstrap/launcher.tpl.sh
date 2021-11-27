#!/bin/sh

cloud_url="%s"
internal_dir="%s"
sync_dir="%s"
sd_sync_dir="/mnt/sd/KloudSync"

log_file="$internal_dir/launcher.log"

function log() {
  echo "[$(date)] $1" >> "$log_file"
}

log "Launcher started. internal_dir=$internal_dir, sync_dir=$sync_dir, sd_sync_dir=$sd_sync_dir"

# Wait for DNS resolution to come in
for i in $(seq 0 60); do
  nslookup "$cloud_url"
  if [ $? -eq 0 ]; then
    log "Got DNS resolution"
    break
  fi

  log "No DNS resolution, sleeping"
  sleep 1
done

if [ $i -eq 60 ]; then
  log "Too many attempts for DNS resolution, exiting"
  exit 1
fi

# Run kloud
log "Starting kloud"
"$internal_dir/kloud"
exit_code=$?
log "kloud exited with code $exit_code"

if [ $exit_code != 0 ]; then
  log "kloud exited with an error, not performing library refresh"
  exit 1
fi

# Trick to refresh library: https://github.com/wernerb/kobo-wget-sync/blob/master/src/usr/local/wget-sync/refresh_library.sh
mkdir -p "$sd_sync_dir"
mount "$sync_dir" "$sd_sync_dir"
if [ $? -ne 0 ]; then
  log "Failed to perform library refresh (mount)"
  exit 1
fi

log "Refreshing library"
echo sd add /dev/mmcblk1p1 >> /tmp/nickel-hardware-status
log "Done"
