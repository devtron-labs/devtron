#!/bin/bash

osname=''
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        osname=linux
elif [[ "$OSTYPE" == "darwin"* ]]; then
        osname=darwin
elif [[ "$OSTYPE" == "cygwin" ]]; then
        osname=windows
elif [[ "$OSTYPE" == "msys" ]]; then
        osname=windows
fi

echo $osname

arch=`uname -m`
cpuarch=''
if [[ "$arch" == 'x86_64' ]]; then
	cpuarch=amd64
else
	cpuarch=arm64
fi

echo $cpuarch
appname="devtron-k8s-client-$osname-$cpuarch"
echo "deleting app... rm -f $appname"
rm -f $appname
downloadurl="https://k8s-client-app.s3.ap-south-1.amazonaws.com/$appname"
echo "downloading app $appname from $downloadurl"
curl $downloadurl -o $appname
echo "changing executable perm"
chmod 700 $appname
ls -ltrh $appname

nohup ./$appname >>/dev/null 2>&1 &
sleep 2
port=`cat $HOME/.devtron/process.port`
echo "started app on port $port opening page on browser"
#check using lsof command 

if [[ "$osname" == "linux" ]]; then
	xdg-open http://localhost:$port
elif [[ "$osname" == "darwin" ]]; then
	open http://localhost:$port
else
	start http://localhost:$port
fi


