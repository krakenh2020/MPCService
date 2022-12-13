#!/bin/bash
# This outputs the file mpc.wasm
GOOS=js GOARCH=wasm go build -o ../manager/assets/mpc.wasm mpc_wasm.go
