#!/bin/bash

PT="json"
SERVER="offline"
AXDBMANAGER="/srv/axihome/bin/axdbManager"

if [ ! -f $AXDBMANAGER ];
then
    AXDBMANAGER="bin/axdbManager"
fi

if [ $# -eq 1 ]
then
    PT=$1
elif [ $# -eq 2 ]
then
    PT=$2
    SERVER=$1
else 
    echo "Bad argument count"
    echo "Usage : $0 serverip(optional) path"
fi

function processOnline {

    $AXDBMANAGER -a set -b $1 -f $PT/$1.json -s $SERVER -p 3330
}

function processOffline {

    $AXDBMANAGER -a set -b $1 -db config/axihome.db -f $PT/$1.json    

    if [ $? -ne 0 ]; then
        echo "-----------------------------------------------------------"
        echo " [Failed : $1]"
        echo "-----------------------------------------------------------"
        exit
    fi
}

function process {

    if [ "$SERVER" == "offline" ];
    then
        processOffline $1
    else
        processOnline $1
    fi    
}

mkdir config  2>/dev/null

for f in $PT*.json
do
    FL=${f##$PT}
    process ${FL%.json}
done
