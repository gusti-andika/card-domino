#!/bin/bash

#for latest protoc
protoc domino.proto --go_out=./ --go-grpc_out=./

#for older protoc
protoc domino.proto --go_out=plugins=grpc:./
