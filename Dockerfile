FROM golang:1.11-stretch as builder

ADD . /src
WORKDIR /src

ENV CGO_ENABLED=1
RUN go build -mod vendor -o exo -ldflags "-w"


FROM ubuntu:bionic
COPY --from=builder /src/exo /
ENTRYPOINT ["/exo"]
