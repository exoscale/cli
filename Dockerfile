FROM golang:1.11-stretch as builder

ADD . /src
WORKDIR /src

ARG VCS_REF
ARG BUILD_DATE

ENV CGO_ENABLED=1
RUN go build -mod vendor -o exo \
        -ldflags "-s -w -X main.version=${BUILD_DATE} -X main.commit=${VCS_REF}"

LABEL org.label-schema.build-date=${BUILD_DATE} \
      org.label-schema.vcs-ref=${VCS_REF} \
      org.label-schema.name="Exo" \
      org.label-schema.vendor="Exoscale" \
      org.label-schema.description="Exoscale CLI" \
      org.label-schema.url="https://github.com/exoscale/cli" \
      org.label-schema.schema-version="1.0"


FROM ubuntu:bionic
COPY --from=builder /src/exo /
ENTRYPOINT ["/exo"]
