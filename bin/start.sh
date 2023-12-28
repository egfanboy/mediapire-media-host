#!/bin/bash
source $PWD/bin/env.sh

if [ -z "$MEDIA_HOST_CONFIG_PATH" ];
then
    echo "MEDIA_HOST_CONFIG_PATH not defined, please add it to the bin/env.sh file and retry again."
    exit 1
fi

if [ -z "$MEDIA_HOST_LOG_PATH" ];
then
    echo "MEDIA_HOST_LOG_PATH not defined, please add it to the bin/env.sh file and retry again."
    exit 1
fi

$PWD/bin/build.sh
echo "Starting service"
nohup $PWD/dist/mediapire-media-host --config $MEDIA_HOST_CONFIG_PATH &> $MEDIA_HOST_LOG_PATH/logs.txt &
echo "Service started with PID $!"
echo $! > $PWD/bin/.pid