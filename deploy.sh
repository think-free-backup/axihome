#!/bin/bash

ssh pi@$1 "sudo systemctl stop axihome"
scp -r bin-arm/* pi@$1:/srv/axihome/bin/
scp -r axihome.assets pi@$1:/srv/axihome/
scp src/github.com/think-free/axihome/importer.sh pi@$1:/srv/axihome/
ssh pi@$1 "sudo systemctl start axihome"

