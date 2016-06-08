#!/bin/bash


echo ""
echo -e "\E[94mAXIHOME\033[0m"
echo -e "\E[94m------------------------------------\033[0m"
echo ""

DEPLOYPATH=deploy/axihome

rm -rf DEPLOYPATH
mkdir -p $DEPLOYPATH  2>/dev/null

echo ""
echo -e "\E[94mGenerating assets\033[0m"
echo ""

mkdir $DEPLOYPATH/assets  2>/dev/null
cd src/github.com/think-free/axihome/
zip -r axihome.assets assets/
mv axihome.assets ../../../../$DEPLOYPATH/assets
cd ../../../../

echo ""
echo -e "\E[94mDeploying binaries\033[0m"
echo ""

mkdir -p $DEPLOYPATH/bin  2>/dev/null
cp bin-arm/axihome/* $DEPLOYPATH/bin/

echo ""
echo -e "\E[94mDeploying service\033[0m"
echo ""

cp src/github.com/think-free/axihome/axihome.service $DEPLOYPATH/

echo ""
echo -e "\E[94mDeploying importer\033[0m"
echo ""

cp src/github.com/think-free/axihome/importer.sh $DEPLOYPATH/

echo ""
echo -e "\E[94mDeploying install script\033[0m"
echo ""

cp src/github.com/think-free/axihome/install.sh $DEPLOYPATH/

echo ""
echo -e "\E[94mGenerate .run\033[0m"
echo ""

cd deploy
chmod +x axihome/install.sh
makeself axihome axihome.`date +%Y.%m.%d`.run "Axihome automation" ./install.sh
cd ..

echo ""
echo -e "\E[94mDone\033[0m"
echo ""

echo ""
echo -e "\E[94mVOICENOTIFIER\033[0m"
echo -e "\E[94m------------------------------------\033[0m"
echo ""

DEPLOYPATH=deploy/voicenotifier

rm -rf DEPLOYPATH
mkdir -p $DEPLOYPATH  2>/dev/null

echo ""
echo -e "\E[94mDeploying binaries\033[0m"
echo ""

mkdir -p $DEPLOYPATH/bin  2>/dev/null
cp bin-arm/voicenotifier/* $DEPLOYPATH/

echo ""
echo -e "\E[94mDeploying application\033[0m"
echo ""

cp src/github.com/think-free/axihome/apps/voicenotifier/svox-pico.tar.gz $DEPLOYPATH/
cp src/github.com/think-free/axihome/apps/voicenotifier/voicenotifier.service $DEPLOYPATH/
cp src/github.com/think-free/axihome/apps/voicenotifier/install.sh $DEPLOYPATH/

echo ""
echo -e "\E[94mGenerate .run\033[0m"
echo ""

cd deploy
chmod +x voicenotifier/install.sh
makeself voicenotifier voicenotifier.`date +%Y.%m.%d`.run "Voice Notifier" ./install.sh
cd ..
