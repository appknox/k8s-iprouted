FROM golang:1.9.3-alpine3.7

WORKDIR /go/src/app

copy . .
