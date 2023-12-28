#!/bin/bash
echo "Deleting dist directory"
rm -rf $PWD/dist
echo "Building service"
go build -o $PWD/dist/mediapire-media-host cmd/main.go