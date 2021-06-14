FROM golang:1.16-alpine as builder

ADD . /src
WORKDIR /src

ARG VERSION
ARG VCS_REF

ENV CGO_ENABLED=0
RUN go build -a -mod vendor -o exo \
        -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${VCS_REF}"

FROM alpine:3.12

WORKDIR /

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
      org.label-schema.url="https://github.com/exoscale/cli" \
      org.label-schema.schema-version="1.0"

RUN apk add --no-cache ca-certificates

COPY --from=builder /src/exo /
ENTRYPOINT ["/exo"]
