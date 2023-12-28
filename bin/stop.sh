#!/bin/bash
PID_FILE=$PWD/bin/.pid
PID=$(cat $PID_FILE)

if kill $PID; then
    echo "Mediapire media host successfully stopped"
    rm $PID_FILE
fi
