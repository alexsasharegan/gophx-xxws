#!/usr/bin/env bash

statik -src www/build
env GOOS=linux GOARCH=arm GOARM=5 go build -o ./pi-ws main.go
