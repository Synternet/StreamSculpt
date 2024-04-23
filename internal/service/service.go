package service

import (
	"bytes"
	"context"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	svcn "streamsculpt/pkg/nats"
	types "streamsculpt/pkg/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"
)

//go:embed abi/*.json
var abiFiles embed.FS

//go:embed abi/map/*.json
var abiMapFiles embed.FS

type Token struct {
	Ticker   string
	Decimals *big.Int
}

type SubscriberConfig struct{}

type PublisherConfig struct {
	SubjectPrefix string
}

type Message struct {
	Postfix string
	Msg     types.DecodedEthLogEvent
}

type MessageChannel chan Message

type SubscriberService struct {
	abis    map[string]abi.ABI
	ctx     context.Context
	cfg     SubscriberConfig
	nats    *svcn.NatsService
	msgChan MessageChannel
}

type PublisherService struct {
	ctx     context.Context
	cfg     PublisherConfig
	nats    *svcn.NatsService
	msgChan MessageChannel
}

func NewSubscriberService(s *svcn.NatsService, ctx context.Context, cfg SubscriberConfig, msgChan MessageChannel) *SubscriberService {
	abis := make(map[string]abi.ABI)

	dirEntries, _ := abiFiles.ReadDir("abi")
	for _, entry := range dirEntries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			data, err := abiFiles.ReadFile("abi/" + entry.Name())
			if err != nil {
				log.Fatalf("failed to read ABI file %s: %v", entry.Name(), err)
			}
			ABI, err := abi.JSON(bytes.NewReader(data))
			log.Printf("Loaded ABI: %s", entry.Name())
			if err != nil {
				log.Fatalf("failed to parse ABI in file %s: %v", entry.Name(), err)
			}
			filenameWithoutExtension := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			abis[filenameWithoutExtension] = ABI
		}
	}

	mapDirEntries, _ := abiMapFiles.ReadDir("abi/map")
	for _, entry := range mapDirEntries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			data, err := abiMapFiles.ReadFile("abi/map/" + entry.Name())
			if err != nil {
				log.Fatalf("failed to read mapping file %s: %v", entry.Name(), err)
			}
			filenameWithoutExtension := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			var mapping []string
			json.Unmarshal(data, &mapping)
			for _, val := range mapping {
				if _, ok := abis[filenameWithoutExtension]; ok {
					abis[val] = abis[filenameWithoutExtension]
					log.Printf("Mapped %s to %s ABI", val, filenameWithoutExtension)
				}
			}
		}
	}

	return &SubscriberService{
		ctx:     ctx,
		cfg:     cfg,
		nats:    s,
		abis:    abis,
		msgChan: msgChan,
	}
}

func NewPublisherService(s *svcn.NatsService, ctx context.Context, cfg PublisherConfig, msgChan MessageChannel) *PublisherService {
	return &PublisherService{
		ctx:  ctx,
		cfg:  cfg,
		nats: s,
		// abi:     ABI,
		msgChan: msgChan,
	}
}

func (s SubscriberService) ProcessTxLogEventFromStream(data []byte) error {
	incoming := types.EthLogEvent{}
	err := json.Unmarshal(data, &incoming)
	if err != nil {
		return err
	}

	abi, ok := s.abis[incoming.Address]
	if !ok {
		return nil
	}

	if err != nil {
		// Not an exhaustive events list. Silently ignore unknown.
		return nil
	}

	eventData, err := hex.DecodeString(strings.TrimPrefix(incoming.Data, "0x"))
	if err != nil {
		log.Fatalf("failed to decode log data: %v", err)
		return nil
	}
	log.Println(string(data))

	outgoing := types.DecodedEthLogEvent{}
	outgoing.Address = incoming.Address
	outgoing.Topics = make([]string, len(incoming.Topics))
	copy(outgoing.Topics, incoming.Topics)
	for i := 1; i < len(incoming.Topics); i++ {
		outgoing.Topics[i] = common.BytesToAddress(common.HexToHash(incoming.Topics[i]).Bytes()).String()
	}
	outgoing.BlockNumber = incoming.BlockNumber
	outgoing.TransactionHash = incoming.TransactionHash
	outgoing.TransactionIndex = incoming.TransactionIndex
	outgoing.BlockHash = incoming.BlockHash
	outgoing.LogIndex = incoming.LogIndex
	outgoing.Removed = incoming.Removed

	event, err := abi.EventByID(common.HexToHash(outgoing.Topics[0]))
	if err != nil {
		log.Fatalf("failed to get event by ID: %v", err)
	}

	outgoing.Data = make(map[string]interface{})
	err = abi.UnpackIntoMap(outgoing.Data, event.Name, eventData)
	if err != nil {
		log.Fatalf("failed to decode %s event log: %v", event.Name, err)
	}
	outgoing.Sig = event.Sig
	s.msgChan <- Message{
		Postfix: fmt.Sprintf(".%s.%s", incoming.Address, event.Name),
		Msg:     outgoing,
	}

	return nil
}

func (s SubscriberService) Serve() {
	serveCtx, cancelFn := context.WithCancel(s.ctx)
	defer cancelFn()

	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Subscriber service is interrupted")
		cancelFn()
	}()

	rungroup, groupCtx := errgroup.WithContext(serveCtx)

	if s.nats != nil {
		rungroup.Go(func() error {
			return s.nats.Serve(groupCtx)
		})
	}

	log.Println("Subscriber service is started")

	if err := rungroup.Wait(); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Printf("Subscriber service is stopped %s", err.Error())
		}
	}

	var completionGroup errgroup.Group
	if s.nats != nil {
		completionGroup.Go(func() error {
			return nil
		})
	}
}

func (s PublisherService) Serve() {
	serveCtx, cancelFn := context.WithCancel(s.ctx)
	defer cancelFn()

	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Publisher service is interrupted")
		cancelFn()
	}()

	rungroup, groupCtx := errgroup.WithContext(serveCtx)

	rungroup.Go(func() error {
		for {
			select {
			case msg := <-s.msgChan:
				subject := fmt.Sprintf("%s%s", s.cfg.SubjectPrefix, msg.Postfix)
				if err := s.nats.PublishAsJSON(groupCtx, subject, msg.Msg); err != nil {
					log.Println(err.Error())
					return err
				}
			case <-groupCtx.Done():
				return groupCtx.Err()
			}
		}
	})

	if s.nats != nil {
		rungroup.Go(func() error {
			return s.nats.Serve(groupCtx)
		})
	}

	log.Println("Publisher service is started")

	if err := rungroup.Wait(); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Printf("Service is stopped %s", err.Error())
		}
	}

	var completionGroup errgroup.Group
	if s.nats != nil {
		completionGroup.Go(func() error {
			return nil
		})
	}
}
