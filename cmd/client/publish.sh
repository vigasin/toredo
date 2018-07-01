#!/bin/bash

set -e

GOOS=linux GOARCH=arm GOARM=7 go build
scp transferer bpi:
