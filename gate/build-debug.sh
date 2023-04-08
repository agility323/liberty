#!/bin/bash

PROC_NAME=gate

go mod tidy

go build -trimpath  -gcflags "all=-N -l" -tags debug -o $PROC_NAME  ./
