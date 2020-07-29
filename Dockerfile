FROM golang:1.14.6

WORKDIR /auth
ADD cmd/ cmd/
ADD pkg/ pkg/
ADD go.mod 
ADD go.sum


# run the server as entrypoint
ENTRYPOINT [ "go", "run", "cmd/server" ]