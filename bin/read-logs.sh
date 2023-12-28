#!/bin/bash
source ./bin/env.sh
if [ -z "$MEDIA_HOST_LOG_PATH" ];
then
    echo "MEDIA_HOST_LOG_PATH not defined, please add it to the bin/env.sh file and retry again."
    exit 1
fi

tail -f -n 1000 $MEDIA_HOST_LOG_PATH/logs.txt