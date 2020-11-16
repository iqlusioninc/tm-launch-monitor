package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	cstypes "github.com/tendermint/tendermint/consensus/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/rpc/client/http"
	tmtypes "github.com/tendermint/tendermint/types"
)

const genesis = "/Users/zakimanian/cosmoshub-test-stargate/config/genesis.json"

const nodeAddress = "http://34.66.219.254:26657"

type roundVotes struct {
	Round              int32    `json:"round"`
	Prevotes           []string `json:"prevotes"`
	PrevotesBitArray   string   `json:"prevotes_bit_array"`
	Precommits         []string `json:"precommits"`
	PrecommitsBitArray string   `json:"precommits_bit_array"`
}

func main() {

	jsonBlob, err := ioutil.ReadFile(genesis)

	if err != nil {
		log.Fatal(err)
	}

	_, err = tmtypes.GenesisDocFromJSON(jsonBlob)
	if err != nil {
		log.Fatal(err)
	}

	c, err := http.New(nodeAddress, "/websocket")
	if err != nil {
		log.Fatalf("can't connect to node: %s", err)
	}

	stateJSON, err := c.ConsensusState(context.Background())
	if err != nil {
		log.Fatalf("can't get consensus state from node: %s", err)
	}

	var state cstypes.RoundStateSimple
	var votes []roundVotes
	err = tmjson.Unmarshal(stateJSON.RoundState, &state)
	if err != nil {
		log.Fatalf("Unmarshalling round state error: %s", err)
	}
	err = tmjson.Unmarshal(state.Votes, &votes)
	if err != nil {
		log.Fatalf("Unmarshalling vote state error: %s", err)
	}

	for _, round := range votes {
		var prevotes []Vote

		var precommits []Vote

		for _, prevoteStr := range round.Prevotes {
			if prevoteStr != "nil-Vote" {
				prevotes = append(prevotes, unpackVote(prevoteStr))
			}
		}

		for _, precommitStr := range round.Precommits {
			if precommitStr != "nil-Vote" {
				precommits = append(precommits, unpackVote(precommitStr))
			}
		}

		fmt.Println(prevotes)

	}

}

type Vote struct {
	position    int64
	fingerprint string
	block       string
	time        time.Time
}

func unpackVote(vote string) Vote {
	var v Vote

	preFixedRemoved := strings.ReplaceAll(vote, "Vote{", "")

	suffixFixedRemoved := strings.ReplaceAll(preFixedRemoved, "}", "")

	components := strings.Split(suffixFixedRemoved, " ")

	fingerprintComponents := strings.Split(components[0], ":")

	position, err := strconv.ParseInt(fingerprintComponents[0], 10, 64)

	if err != nil {
		log.Fatalf("Cannot parse position from %s", vote)
	}

	v.position = int64(position)
	v.fingerprint = fingerprintComponents[1]
	v.block = components[2]

	voteTime, err := time.Parse(time.RFC3339, components[5])
	if err != nil {
		log.Fatalf("Cannot parse position from %s", vote)
	}
	v.time = voteTime

	return v
}
