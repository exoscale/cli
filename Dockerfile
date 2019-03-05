FROM golang:1.12-stretch as builder

ADD . /src
WORKDIR /src

ARG VERSION
ARG VCS_REF

ENV CGO_ENABLED=1
RUN go build -mod vendor -o exo \
        -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${VCS_REF}"

FROM ubuntu:cosmic

ARG VERSION
ARG VCS_REF
ARG BUILD_DATE

LABEL org.label-schema.build-date=${BUILD_DATE} \
      org.label-schema.vcs-ref=${VCS_REF} \
      org.label-schema.vcs-url="https://github.com/exoscale/cli" \
      org.label-schema.version=${VERSION} \
      org.label-schema.name="exo" \
      org.label-schema.vendor="Exoscale" \
      org.label-schema.description="Exoscale CLI" \
      org.label-schema.url="https://exoscale.github.io/cli" \
      org.label-schema.schema-version="1.0"

RUN set -xe \
 && apt-get update -q \
 && apt-get upgrade -q -y \
 && apt-get install -q -y \
        ca-certificates \
 && apt-get autoremove -y \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /src/exo /
ENTRYPOINT ["/exo"]
