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
	"path/filepath"
	"strings"

	types "github.com/synternet/StreamSculpt/pkg/types"
	"github.com/synternet/data-layer-sdk/pkg/options"
	"github.com/synternet/data-layer-sdk/pkg/service"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

//go:embed abi/*.json
var abiFiles embed.FS

//go:embed abi/map/*.json
var abiMapFiles embed.FS

type Token struct {
	Ticker   string
	Decimals *big.Int
}

type Message struct {
	Postfix string
	Msg     types.DecodedEthLogEvent
}

type MessageChannel chan Message

type Publisher struct {
	*service.Service
	msgChan MessageChannel
	abis    map[string]abi.ABI
}

func New(opts ...options.Option) *Publisher {
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

	ret := &Publisher{
		Service: &service.Service{},
		msgChan: make(MessageChannel, 1024),
		abis:    abis,
	}

	ret.Configure(opts...)

	return ret
}

func (p *Publisher) subscribe() error {
	src := options.Param(p.Options, "TxLogEventsStreamSubject", "")
	if src == "" {
		return errors.New("source subject must not be empty")
	}
	if _, err := p.SubscribeTo(p.handleTxLogEvent, src); err != nil {
		return err
	}

	return nil
}

func (p *Publisher) handleTxLogEvent(nmsg service.Message) {
	incoming := types.EthLogEvent{}
	err := json.Unmarshal(nmsg.Data(), &incoming)
	if err != nil {
		fmt.Errorf("Failed to decode JSON: %v", err.Error())
		return
		// return err
	}

	abi, ok := p.abis[incoming.Address]
	if !ok {
		// fmt.Errorf("Failed to find ABI: %v", err.Error())
		return
		// return nil
	}

	if err != nil {
		// Not an exhaustive events list. Silently ignore unknown.
		return
		// return nil
	}

	eventData, err := hex.DecodeString(strings.TrimPrefix(incoming.Data, "0x"))
	if err != nil {
		log.Fatalf("failed to decode log data: %v", err)
		return
		// return nil
	}
	log.Println(string(nmsg.Data()))

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

	// s.msgChan <- Message{
	// 	Postfix: fmt.Sprintf(".%s.%s", incoming.Address, event.Name),
	// 	Msg:     outgoing,
	// }

	postfix := fmt.Sprintf("%s.%s", incoming.Address, event.Name)
	// src := options.Param(p.Options, "", "")
	// subject := fmt.Sprintf("%s%s", src, postfix)

	fmt.Println("loggged")
	p.PublishBuf(outgoing.AsJSON(), postfix)
	// subject := fmt.Sprintf("%s%s", s.cfg.SubjectPrefix, postfix)
	// if err := p.nats.PublishAsJSON(groupCtx, subject, msg.Msg); err != nil {
	// 	log.Println(err.Error())
	// 	return err
	// }
}

func (s *Publisher) Start() context.Context {
	err := s.subscribe()
	if err != nil {
		s.Fail(err)
		return s.Context
	}

	return s.Service.Start()
}
