# WASM files for interface with JS

This folder includes all needed to compile Go code with WASM to be able to run it in JS. The interface
for requesting MPC computation is implemented in as a web service and needed Go code can be run using WASM.

### Build
To compile a .wasm file that provides functions needed

``GOOS=js GOARCH=wasm go build -o ../manager/assets/mpc.wasm mpc_wasm.go``

this will put the wasm file to managers assets, which will be served to a web page in the browser.
