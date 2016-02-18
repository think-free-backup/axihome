#!/bin/bash

GOA="amd64"

if [ $# -eq 1 ]
then
    
    export GOBIN="$GOPATH/bin-$1/"
    mkdir $GOPATH/bin-$1/ 2>/dev/null
    GOA=$1
fi

function build(){

    echo "- Building $2 ($1)"
    GOARCH=$1 go install src/github.com/think-free/axihome/$2
}

echo ""
echo -e "\E[94mBuilding for arch $GOA\033[0m"
echo ""

build $GOA backends/v2bridge/v2bridge.go
build $GOA backends/upsc/upsc.go
build $GOA backends/pfsensegwstate/pfsensegwstate.go
build $GOA backends/ping/ping.go
build $GOA backends/virtual/virtual.go
build $GOA backends/time/time.go
build $GOA core/notification/notification.go
build $GOA core/variablesnotification/variablesnotification.go
build $GOA core/variablebinding/variablebinding.go
build $GOA core/variablecalculation/variablecalculation.go
build $GOA core/wearablegw/wearablegw.go
build $GOA tools/axdbManager.go
build $GOA tools/axVariableWriter.go
build $GOA axihome.go

export GOBIN="$GOPATH/bin/"

echo ""
echo -e "\E[94mGenerating assets\033[0m"
echo ""

cd src/github.com/think-free/axihome/
zip -r axihome.assets assets/
mv axihome.assets ../../../../
cd ../../../../

echo ""
echo -e "\E[94mDone\033[0m"
echo ""
