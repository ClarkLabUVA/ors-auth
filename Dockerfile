FROM golang:1.14.6 as builder

WORKDIR /auth
ADD cmd/ cmd/
ADD pkg/ pkg/
ADD go.mod go.mod

# run the server as entrypoint
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go 

FROM alpine:latest

WORKDIR /auth
COPY --from=builder /auth/server .

ENTRYPOINT [ "./server" ]