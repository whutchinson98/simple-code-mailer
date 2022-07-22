#!/bin/sh

cd app/emailSender
GOOS=linux GOARCH=amd64 go build -o emailSender .

cd ../emailGenerator
GOOS=linux GOARCH=amd64 go build -o emailGenerator .
