#!/bin/sh

# build the binary
env GOOS=linux GOARCH=amd64 go build -o bin/sample-serverless-api .

# compress the binary
zip -j main.zip bin/sample-serverless-api

# update lambda function
aws lambda update-function-code --function-name books --zip-file fileb://main.zip
