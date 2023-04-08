#!/bin/bash

PROC_NAME=gate

go mod tidy

go build -trimpath  -gcflags "all=-N -l" -o $PROC_NAME  ./
