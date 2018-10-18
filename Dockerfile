FROM golang:1.11-stretch as builder

ADD . /src
WORKDIR /src

ARG VCS_REF
ENV ENV_VCS_REF=$VCS_REF

ARG BUILD_DATE
ENV ENV_BUILD_DATE=$BUILD_DATE

ENV CGO_ENABLED=1
RUN go build -mod vendor -o exo \
        -ldflags "-s -w -X main.version=${ENV_BUILD_DATE} -X main.commit=${ENV_VCS_REF}"

FROM ubuntu:bionic
COPY --from=builder /src/exo /
ENTRYPOINT ["/exo"]
