# Build Container
FROM golang:1.18 AS build

ARG VERSION
ARG NATS_URLS NATS_SUB_CREDS NATS_PUB_CREDS NATS_RECONNECT_WAIT NATS_MAX_RECONNECT NATS_EVENT_LOG_STREAM_SUBJECT
ARG NATS_UNPACKED_STREAMS_SUBJECT_PREFIX

COPY . /home/src
WORKDIR /home/src

RUN CGO_ENABLED=0 go build -v -o . -installsuffix cgo -ldflags="-w -s" ./...

# Runtime Container
FROM alpine

ENV PATH="/app:${PATH}"

RUN apk add --update --no-cache \
    ca-certificates \
    bash

WORKDIR /home/app
COPY --from=build /home/src/streamsculpt .
COPY --from=build /home/src/docker/entrypoint.sh /home/app
RUN chmod +x /home/app/entrypoint.sh

ENTRYPOINT ["/home/app/entrypoint.sh"]
