package main

import (
	"context"
	"flag"
	"log"
	"time"

	"streamsculpt/internal/service"

	svcnats "streamsculpt/pkg/nats"

	nats "github.com/nats-io/nats.go"
)

func main() {
	flagNatsUrls := flag.String("nats-urls", "nats://34.107.87.29 ", "NATS server URLs (separated by comma)")
	flagUserCredsSeedSub := flag.String("nats-sub-nkey", "", "NATS subscriber user credentials NKey string")
	flagUserCredsSeedPub := flag.String("nats-pub-nkey", "", "NATS publisher user credentials NKey string")
	flagNatsReconnectWait := flag.Duration("nats-reconnect-wait", 10*time.Second, "NATS reconnect wait duration")
	flagNatsMaxReconnects := flag.Int("nats-max-reconnect", 500, "NATS max reconnect attempts count")
	flagNatsTxLogEventsStreamSubject := flag.String("nats-event-log-stream-subject", "syntropy.ethereum.log-event", "NATS event log stream subject")
	flagNatsUnpackedStreamsSubjectPrefix := flag.String("nats-unpacked-streams-subject-prefix", "", "NATS event log stream subject")

	flag.Parse()

	if flagNatsUnpackedStreamsSubjectPrefix == nil || *flagNatsUnpackedStreamsSubjectPrefix == "" {
		log.Fatalf("missing flag: nats-unpacked-streams-subject-prefix")
	}

	optsSub := []nats.Option{}

	flagUserCredsJWTSub, err := svcnats.CreateAppJwt(*flagUserCredsSeedSub)
	if err != nil {
		log.Fatalf("failed to create sub JWT: %v", err)
	}
	optsSub = append(optsSub, nats.UserJWTAndSeed(flagUserCredsJWTSub, *flagUserCredsSeedSub))

	optsSub = append(optsSub, nats.MaxReconnects(*flagNatsMaxReconnects))
	optsSub = append(optsSub, nats.ReconnectWait(*flagNatsReconnectWait))

	optsPub := []nats.Option{}

	flagUserCredsJWTPub, err := svcnats.CreateAppJwt(*flagUserCredsSeedPub)
	if err != nil {
		log.Fatalf("failed to create pub JWT: %v", err)
	}
	optsPub = append(optsPub, nats.UserJWTAndSeed(flagUserCredsJWTPub, *flagUserCredsSeedPub))

	optsPub = append(optsPub, nats.MaxReconnects(*flagNatsMaxReconnects))
	optsPub = append(optsPub, nats.ReconnectWait(*flagNatsReconnectWait))

	svcnSub := svcnats.MustConnect(
		svcnats.Config{
			URI:  *flagNatsUrls,
			Opts: optsSub,
		})
	log.Println("NATS sub service connected")

	svcnPub := svcnats.MustConnect(
		svcnats.Config{
			URI:  *flagNatsUrls,
			Opts: optsPub,
		})
	log.Println("NATS pub service connected")

	txMsgChannel := make(service.MessageChannel, 1024)

	cfgSub := service.SubscriberConfig{}
	cfgPub := service.PublisherConfig{
		SubjectPrefix: *flagNatsUnpackedStreamsSubjectPrefix,
	}
	sSub := service.NewSubscriberService(svcnSub, context.Background(), cfgSub, txMsgChannel)
	sPub := service.NewPublisherService(svcnPub, context.Background(), cfgPub, txMsgChannel)

	svcnSub.AddHandler(*flagNatsTxLogEventsStreamSubject, sSub.ProcessTxLogEventFromStream)

	go sPub.Serve()
	sSub.Serve()
}
