package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	service "github.com/synternet/StreamSculpt/internal/service"
	"github.com/synternet/data-layer-sdk/pkg/options"
	svc "github.com/synternet/data-layer-sdk/pkg/service"
	"github.com/synternet/data-layer-sdk/pkg/user"
)

func main() {
	flagNatsUrls := flag.String("nats-urls", "nats://34.107.87.29 ", "NATS server URLs (separated by comma)")
	flagUserCredsSeedSub := flag.String("nats-sub-nkey", "", "NATS subscriber user credentials NKey string")
	flagUserCredsSeedPub := flag.String("nats-pub-nkey", "", "NATS publisher user credentials NKey string")
	flagNatsTxLogEventsStreamSubject := flag.String("nats-event-log-stream-subject", "synternet.ethereum.log-event", "NATS event log stream subject")
	flagNatsPubPrefix := flag.String("nats-pub-prefix", "synternet", "NATS event log stream prefix")
	flagNatsPubName := flag.String("nats-pub-name", "ethereum.unpacked", "NATS event log stream name")

	flag.Parse()

	userCredsSeedSub, userCredsJWTSub, err := user.CreateCreds([]byte(*flagUserCredsSeedSub))
	if err != nil {
		log.Fatalf("failed to create sub JWT: %v", err)
	}

	connSub, err := options.MakeNats("Streaming consumer", *flagNatsUrls, "", userCredsSeedSub, userCredsJWTSub, "", "", "")
	if err != nil {
		panic(fmt.Errorf("Failed creating NATS connection: %w", err))
	}

	userCredsSeedPub, userCredsJWTPub, err := user.CreateCreds([]byte(*flagUserCredsSeedPub))
	if err != nil {
		log.Fatalf("failed to create pub JWT: %v", err)
	}

	connPub, err := options.MakeNats("Streaming producer", *flagNatsUrls, "", userCredsSeedPub, userCredsJWTPub, "", "", "")
	if err != nil {
		panic(fmt.Errorf("Failed creating NATS connection: %w", err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	opts := []options.Option{
		svc.WithContext(ctx),
		svc.WithSubNats(connSub),
		svc.WithPubNats(connPub),
		svc.WithPrefix(*flagNatsPubPrefix),
		svc.WithName(*flagNatsPubName),
		svc.WithParam("TxLogEventsStreamSubject", *flagNatsTxLogEventsStreamSubject),
	}

	s := service.New(opts...)
	defer s.Close()

	pubCtx := s.Start()

	select {
	case <-ctx.Done():
		log.Println("Shutdown")
	case <-pubCtx.Done():
		log.Println("Publisher stopped with cause: ", context.Cause(pubCtx).Error())
	}
}
