package ethereum

import (
	"fmt"
	"strings"

	types "streamsculpt/pkg/types"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

func hashEventSignature(eventSignature string) string {
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(eventSignature))
	return common.Bytes2Hex(hash.Sum(nil))
}

func GetEventName(e types.EthLogEvent) (string, error) {
	eventSignatures := map[string]string{
		hashEventSignature("Swap(address,uint256,uint256,uint256,uint256,address)"): "Swap",
	}
	if len(e.Topics) == 0 {
		return "", fmt.Errorf("no topics found in the event")
	}

	topic := strings.TrimPrefix(e.Topics[0], "0x")

	eventName, ok := eventSignatures[topic]
	if !ok {
		return "", fmt.Errorf("unknown event topic: %s", topic)
	}

	return eventName, nil
}
