instructions coming sooner or later since this is far away from being finished!

## Build instructions
1. install golang
2. in the root of this project do:
```sh
go get github.com/google/uuid
go get github.com/gorilla/websocket
go build
./trafficker
```
this will start the application on localhost:8080

## setup the echo websocket
```sh
cd tester
npm i
node echo_client.js
```
will connect to localhost:8080/ws
