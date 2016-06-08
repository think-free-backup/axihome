#!/bin/bash

systemctl stop voicenotifier 2>/dev/null

rm -rf /opt/svox-pico 2>/dev/null
tar xf svox-pico.tar.gz 2>/dev/null
mv svox-pico /opt/svox-pico 2>/dev/null
mkdir /srv/voicenotifier
mv voicenotifier /srv/voicenotifier/

mv voicenotifier.service /etc/systemd/system/voicenotifier.service
systemctl enable voicenotifier

chown pi:pi /srv/voicenotifier
chown pi:pi /opt/svox-pico

echo "Start voicenotifier with 'systemctl start voicenotifier'"
