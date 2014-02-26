// very basic markov chains-based text generation
// usage: go run markov.go -file /file/to/train/on.txt -amount 100

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"math"
	"math/rand"
	"strings"
)

var amount int
var path string

func init() {
	flag.IntVar(&amount, "amount", 10, "amount of words to produce")
	flag.StringVar(&path, "file", "example.txt", "file to train the markov chains on")
	flag.Parse()
}



// tokenization function on the file
func split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance, token, err = bufio.ScanWords(data, atEOF)
	if advance > 0 {
		token = []byte(strings.ToLower(strings.Trim(string(token), "_:-,!.?;\"''")))
	}
	return
}

// tokenize the text file
func tokenizeTextFile(filename string, ch chan string) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	r := bufio.NewReader(file)
	scanner := bufio.NewScanner(r)
	scanner.Split(split)

	for scanner.Scan() {
		ch <- scanner.Text()
	}
	close(ch)
}

// generate random numbers on demand and push them in a channel
func randGen(ch chan float64) {
	rand.Seed(31)
	for {
		ch <- rand.Float64()
	}
}

// return an integer in [0;max) from the generator
func intfrom(ch chan float64, max int) int {
	return int(math.Floor((<-ch) * float64(max)))
}

// pick a random neighbor from a markov node
func pickRandom(markov map[string]map[string]int, key string, randChan chan float64) string {
	nodes := markov[key]
	keys := make([]string, 0, len(nodes))
	total := 0
	for key, value := range nodes {
		keys = append(keys, key)
		total += value
	}
	r := intfrom(randChan, total)
	current_key := ""
	for _, current_key = range keys {
		value := nodes[current_key]
		r -= value
		if r < 0 {
			break
		}
	}
	return current_key
}


func main() {
	randChan := make(chan float64, 10000)
	tokenChan := make(chan string)
	go randGen(randChan)
	go tokenizeTextFile(path, tokenChan)

	markov := make(map[string]map[string]int)
	unique_words := make([]string, 0)

	// train markov chain
	previous := ""
	for tok := range tokenChan {
		if len(tok) > 0 {
			if len(previous) > 0 {
				_, ok := markov[previous]
				if !ok {
					markov[previous] = make(map[string]int)
					unique_words = append(unique_words, previous)
				}
				markov[previous][tok] += 1
			}
			previous = tok
		}
	}

	// exploit markov chain
	start := unique_words[intfrom(randChan, len(unique_words))]
	//first := true
	for i := 0; i < amount; i += 1 {
		fmt.Print(start)
		fmt.Print(" ")
		start = pickRandom(markov, start, randChan)
	}
	fmt.Print("\n")
	close(randChan)
}
