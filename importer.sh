#!/bin/bash

PT="json"
SERVER="offline"

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

    bin/axdbManager -a set -b $1 -f $PT/$1.json -s $SERVER -p 3330
}

function processOffline {

    bin/axdbManager -a set -b $1 -db axihome.db -f $PT/$1.json    
}

function process {

    if [ "$SERVER" == "offline" ];
    then
        processOffline $1
    else
        processOnline $1
    fi    
}


process "Variables"
process "Config"
process "Instances"
process "NotificationGeneral"
process "NotificationMessages"
process "NotificationMessagesSubscriptions"
process "NotificationDevices"
process "VariablesNotification"
process "WearableVoiceAction"
process "VariablesBinding"
process "VariablesCalculation"
