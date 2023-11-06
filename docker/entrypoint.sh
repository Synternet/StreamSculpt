#!/bin/sh

CMD="./streamsculpt"

if [ ! -z "$NATS_URLS" ]; then
  CMD="$CMD --nats-urls $NATS_URLS"
fi

if [ ! -z "$NATS_SUB_NKEY" ]; then
  CMD="$CMD --nats-sub-nkey $NATS_SUB_NKEY"
fi

if [ ! -z "$NATS_PUB_NKEY" ]; then
  CMD="$CMD --nats-pub-nkey $NATS_PUB_NKEY"
fi

if [ ! -z "$NATS_RECONNECT_WAIT" ]; then
  CMD="$CMD --nats-reconnect-wait $NATS_RECONNECT_WAIT"
fi

if [ ! -z "$NATS_MAX_RECONNECT" ]; then
  CMD="$CMD --nats-max-reconnect $NATS_MAX_RECONNECT"
fi

if [ ! -z "$NATS_EVENT_LOG_STREAM_SUBJECT" ]; then
  CMD="$CMD --nats-event-log-stream-subject $NATS_EVENT_LOG_STREAM_SUBJECT"
fi

if [ ! -z "$NATS_UNPACKED_STREAMS_SUBJECT_PREFIX" ]; then
  CMD="$CMD --nats-unpacked-streams-subject-prefix $NATS_UNPACKED_STREAMS_SUBJECT_PREFIX"
fi

exec $CMD
