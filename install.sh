#!/bin/bash

systemctl stop axihome 2>/dev/null
mkdir -p /var/log/axihome 2>/dev/null
mkdir -p /etc/axihome/config  2>/dev/null
rm -rf /srv/axihome-old 2>/dev/null
mv /srv/axihome /srv/axihome-old 2>/dev/null
mkdir -p /srv/axihome  2>/dev/null
mv assets /srv/axihome/
mv bin /srv/axihome/

mv axihome.service /etc/systemd/system/axihome.service
systemctl enable axihome

mv importer.sh /etc/axihome/config/
chown -R pi:pi /srv/axihome
chown -R pi:pi /etc/axihome
chown -R pi:pi /var/log/axihome

echo "Import configuration to /etc/axihome/config"
echo "Start axihome with 'systemctl start axihome'"
