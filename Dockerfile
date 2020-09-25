FROM golang:1.15 as builder

WORKDIR /go/src/app
COPY . .

RUN go get -u k8s.io/code-generator/...
RUN CGO_ENABLED=0 APP_BIN_NAME=controller make build

FROM alpine:3.12

WORKDIR /opt

COPY --from=builder /go/src/app/controller ./controller

ENTRYPOINT ["/opt/controller"]
