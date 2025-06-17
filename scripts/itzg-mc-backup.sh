#!/bin/bash

docker run --rm \
  --name mc-backup \
  --network main-application-network2 \
  -v /usr/share/minecraft_mnt/minecraft-data:/data:ro \
  -v /usr/share/minecraft_mnt/mc_backups:/backups \
  -e WORLD_SAVE_NAME=world_ver_1 \
  -e ENABLE_RCON=true \
  -e RCON_HOST=vanilla \
  -e RCON_PORT=25575 \
  -e RCON_PASSWORD=supersecurepassword \
  -e PRUNE_BACKUPS_DAYS=7 \
  -e BACKUP_INTERVAL=0 \
  itzg/mc-backup >> /var/log/mc-backup/backup.log 2>&1