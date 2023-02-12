#!/bin/bash

DEVTRON_BASE_DIR="$HOME/.devtron"
DEVTRON_APP_PORT_PATH="$DEVTRON_BASE_DIR/process.port"
if [[ ! -d "$DEVTRON_BASE_DIR" ]]; then
	mkdir $DEVTRON_BASE_DIR
fi

openBrowser () {
	osname=$(getOsName)
	PROCESS_FILE_PATH=$DEVTRON_APP_PORT_PATH
	if [[ -f "$PROCESS_FILE_PATH" ]]; then
		port=`cat $PROCESS_FILE_PATH`
		if [[ "$osname" == "linux" ]]; then
			xdg-open http://localhost:$port
		elif [[ "$osname" == "darwin" ]]; then
			open http://localhost:$port
		else
			start http://localhost:$port
		fi
	else
		echo "no process running"
	fi
}

getAppPath () {
	echo "$DEVTRON_BASE_DIR/$(getAppName)"
}

checkAppExists () {
	appPath=$(getAppPath)
	if [[ -f "$appPath" ]]; then
		echo 1
	fi
	echo -1
}



getAppProcessId () {
	PROCESS_FILE_PATH=$DEVTRON_APP_PORT_PATH
	if [[ -f "$PROCESS_FILE_PATH" ]]; then
		echo "$PROCESS_FILE_PATH file exists"
		port=`cat $PROCESS_FILE_PATH`
		pid=`lsof -i:$port | grep LISTEN | cut -d' ' -f2 | uniq`
		echo $pid
	else
		echo "-1"
	fi
}

checkAndDownloadApp () {
	appExists=$(checkAppExists)
	if [[ "$appExists" == "-1" ]]; then
		downloadApp
	else
		echo "app already exists at path $(getAppPath)"
	fi
}


startApp () {
	checkAndDownloadApp
	existingPid=$(getAppProcessId)
	if [[ "$existingPid" > "-1" ]]; then
		echo "cannot start a new app already running pid " $existingPid " try open instead of start"
		exit 1
	fi
	appPath=$(getAppPath)
	nohup "$appPath" >>/dev/null 2>&1 &
	sleep 5
	port=`cat $DEVTRON_APP_PORT_PATH`
	echo "started app on port $port opening on browser"
	openBrowser
}

getOsName () {
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
}
getOsArch () {
	arch=`uname -m`
	cpuarch=''
	if [[ "$arch" == 'x86_64' ]]; then
		cpuarch=amd64
	else
		cpuarch=arm64
	fi
	echo $cpuarch
}

getAppName () {
	osname=$(getOsName)
	cpuarch=$(getOsArch)
	appname="devtron-k8s-client-$osname-$cpuarch"
	echo $appname
}

deleteApp () {
  appPath=$(getAppPath)
  echo "deleting app... rm -f $appPath"
  rm -f $appPath
}

downloadApp () {
	appname=$(getAppName)
	appPath=$(getAppPath)
	deleteApp
	downloadurl="https://k8s-client-app.s3.ap-south-1.amazonaws.com/$appname"
	echo "downloading app $appname from $downloadurl"
	curl $downloadurl -o $appPath
	echo "changing executable perm"
	chmod 700 $appPath
	ls -ltrh $appPath
}

checkAndStopApp () {
  if [[ -f "$DEVTRON_APP_PORT_PATH" ]]; then
    echo "stopping app"
    port=`cat $DEVTRON_APP_PORT_PATH`
    if [[ "$port" != "" ]]; then
      pid=`lsof -i:$port | cut -d' ' -f2`
      kill -9 $pid
      rm -f $DEVTRON_APP_PORT_PATH
    fi
  fi
}

############## CHECK STOP CMD ##############
if [[ "$1" == "stop" && -f "$DEVTRON_APP_PORT_PATH" ]]; then
	echo "received "$1
	port=`cat $DEVTRON_APP_PORT_PATH`
	if [[ "$port" == "" ]]; then
		echo "no running process so far"
		exit 1
	fi
	pid=`lsof -i:$port | cut -d' ' -f2`
	kill -9 $pid
	rm -f $DEVTRON_APP_PORT_PATH
	exit 1
elif [[ "$1" == "stop" ]]; then
	echo "received "$1
	echo "no running process so far"
	exit 1
elif [[ "$1" == "start" ]]; then
	echo "received "$1
	startApp
	exit 1
elif [[ "$1" == "open" ]]; then
	echo "received "$1
	openBrowser
	exit 1
elif [[ "$1" == "upgrade" ]]; then
  echo "upgrade received"
  checkAndStopApp
  deleteApp
  startApp
  exit 1
fi

existingPid=$(getAppProcessId)
if [[ "$existingPid" > "-1" ]]; then
	echo "cannot start a new app already running pid " $existingPid " try open instead of start"
	exit 1
fi

checkAndDownloadApp

############## START APP ##############
startApp
