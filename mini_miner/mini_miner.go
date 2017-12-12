package main

import "fmt"
import "bytes"
import "crypto/sha256"
import "time"
import "encoding/json"
import "math/bits"
import "net/http"
import "log"
import "os"

var token string

/*

Returned problem JSON:

{"difficulty": 13, "block": {"nonce": null, "data": [["7777abe2d01e57743988e9667382508f", -30], ["c4cd50ca15e05d074d1ab93af1a3f37f", -16], ["21529503e7faf3c211d3a4b39fde5589", 8], ["40ce32d34ff8883545966eb71556bba5", -10], ["8926acf67695aa8c2cef27fb622bd385", 89]]}}

*/

type Problem struct {
	Difficulty int   `json:"difficulty"`
	Block      Block `json:"block"`
}

type Block struct {
	Data  []interface{} `json:"data"`
	Nonce int           `json:"nonce"`
}

type Solution struct {
	Nonce int `json:"nonce"`
}

func GetProblem() Problem {
	netClient := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := netClient.Get("https://hackattic.com/challenges/mini_miner/problem?access_token=" + token)
	if err != nil {
		log.Fatal("Couldn't get problem:", err)
	}
	decoder := json.NewDecoder(resp.Body)
	problem := Problem{}
	err = decoder.Decode(&problem)
	if err != nil {
		log.Fatal("Error decoding response:", err)
	}
	return problem
}

func SHA256(block Block) [32]byte {
	jsonBuf, err := json.Marshal(&block)
	if err != nil {
		log.Fatal("Couldn't JSON encode block", block, ":", err)
	}
	return sha256.Sum256(jsonBuf)
}

func LeadingZeroBits(buf [32]byte) int {
	zeros := 0
	for _, b := range buf {
		if b == 0 {
			zeros += 8
			continue
		}
		zeros += bits.LeadingZeros8(uint8(b))
		break
	}
	return zeros
}

func SubmitSolution(nonce int) {
	netClient := &http.Client{
		Timeout: time.Second * 10,
	}

	solution := Solution{nonce}
	solutionJson, err := json.Marshal(&solution)
	if err != nil {
		log.Fatal("Couldn't marshal solution", solution, ":", err)
	}
	resp, err := netClient.Post(
		"https://hackattic.com/challenges/mini_miner/solve?access_token="+token,
		"application/json",
		bytes.NewReader(solutionJson),
	)
	if err != nil {
		log.Fatal("Couldn't do POST:", err)
	}

	fmt.Println("Got response to solution:", resp, resp.Body)
}

func main() {
	token = os.Getenv("HACKATTIC_TOKEN")
	if token == "" {
		fmt.Errorf("Run \"export HACKATTIC_TOKEN=your_token_here\"\n")
		os.Exit(1)
	}
	fmt.Println("Using HACKATTIC_TOKEN: ", token)

	problem := GetProblem()
	fmt.Printf("Got problem %+v\n", problem)
	leadingZeros := 0
	for leadingZeros < problem.Difficulty {
		problem.Block.Nonce++
		leadingZeros = LeadingZeroBits(SHA256(problem.Block))
	}
	SubmitSolution(problem.Block.Nonce)
}
