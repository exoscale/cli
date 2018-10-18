FROM golang:1.11-stretch as builder

ADD . /src
WORKDIR /src

ENV CGO_ENABLED=1
RUN go build -mod vendor -o exo \
        -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${VCS_REF}"


FROM ubuntu:bionic
COPY --from=builder /src/exo /
ENTRYPOINT ["/exo"]
