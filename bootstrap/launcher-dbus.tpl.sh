#!/bin/sh

#NOTE: This is an experiment that is currently NOT working

export DBUS_SESSION_BUS_ADDRESS=unix:path=//var/run/dbus/system_bus_socket 

kloud_library_folder="/mnt/onboard/KloudSync"
sd_library_folder="/mnt/sd/KloudSync"

kloud_internal_folder="/mnt/onboard/.kloud"
log_file="$kloud_internal_folder/launcher.log"

function log() {
  echo "[$(date)] --- $1" >> "$log_file"
}

type="signal"
interface="name.marples.roy.dhcpcd"
member="StatusChanged"
watch_exp="type='$type',interface='$interface',member='$member'"

dbus-monitor "$watch_exp" | while read -r line; do
  if echo $line | grep -q '"connected"'; then
    # Make sure we have internet access
    ping -c2 1.1.1.1 > /dev/null

    # Run kloud
    log "Starting kloud"
    "$kloud_internal_folder/kloud"
    exit_code=$?

    log "kloud exited with code $exit_code"
    if [ $exit_code != 0 ]; then
      log "kloud exited with an error, not performing library refresh"
    else
      # Trick to refresh library: https://github.com/wernerb/kobo-wget-sync/blob/master/src/usr/local/wget-sync/refresh_library.sh
      mkdir -p "$sd_library_folder"
      mount "$kloud_library_folder" "$sd_library_folder"

      log "Refreshing library"
      echo sd add /dev/mmcblk1p1 >> /tmp/nickel-hardware-status
      log "Done"
    fi
  fi
done
