#!/bin/bash

VERSION=0.2
ZIPNAME=manager-v${VERSION}.zip

GOOS=linux go build

zip ${ZIPNAME} manager
aws s3 cp ${ZIPNAME} s3://toredo-lambda/manager/${ZIPNAME}
