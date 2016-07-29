#!/bin/bash

GOA="amd64"
GOBINLOC="$GOPATH/bin"
SPECIFIC=0

if [ $# -eq 1 ]
then
    SPECIFIC=1
    export GOBINLOC="$GOPATH/bin-$1/"
    mkdir -p $GOPATH/bin-$1/ 2>/dev/null
    GOA=$1
fi

function build(){

    if [ $SPECIFIC -eq 1 ]
    then
        export GOBIN="$GOBINLOC/$2"
        mkdir $GOBIN 2>/dev/null
    fi

    echo "- Building $3 ($1)"
    GOARCH=$1 go install src/github.com/think-free/axihome/$3 # -ldflags "-s"
    
    export GOBIN=$GOBINLOC
}

echo ""
echo -e "\E[94mBuilding for arch $GOA\033[0m"
echo ""

build $GOA axihome backends/v2bridge/v2bridge.go
build $GOA axihome backends/upsc/upsc.go
build $GOA axihome backends/pfsensegwstate/pfsensegwstate.go
build $GOA axihome backends/ping/ping.go
build $GOA axihome backends/virtual/virtual.go
build $GOA axihome backends/time/time.go
build $GOA axihome backends/hue/hue.go
build $GOA axihome backends/waze/waze.go
build $GOA axihome backends/quote/quote.go
build $GOA axihome backends/weather/weather.go
build $GOA axihome backends/fitbit/fitbit.go
build $GOA axihome backends/ipx800/ipx800.go
build $GOA axihome backends/dgtpanel/dgtpanel.go
build $GOA axihome core/historic/historic.go
build $GOA axihome core/chart/chart.go
build $GOA axihome core/notification/notification.go
build $GOA axihome core/variablesnotification/variablesnotification.go
build $GOA axihome core/variablebinding/variablebinding.go
build $GOA axihome core/variablecalculation/variablecalculation.go
build $GOA axihome core/wearablegw/wearablegw.go
build $GOA axihome core/acond/acond.go
build $GOA axihome core/variabletimer/variabletimer.go
build $GOA axihome core/scheduler/scheduler.go
build $GOA axihome core/tschecker/tschecker.go
build $GOA axihome tools/axdbManager.go
build $GOA axihome tools/axwrite.go
build $GOA axihome tools/axset.go
build $GOA axihome tools/axget.go
build $GOA axihome axihome.go

echo ""
echo -e "\E[94mBuilding tools for arch $GOA\033[0m"
echo ""

build $GOA iot6lowpan apps/iot6lowpan/iot6lowpan.go 
build $GOA voicenotifier apps/voicenotifier/voicenotifier.go

export GOBIN="$GOPATH/bin/"
