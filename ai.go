package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/cdipaolo/goml/base"
	"github.com/cdipaolo/goml/text"
)

var model *text.NaiveBayes

func split(r rune) bool {
	return r == '/' || r == '-' || r == '.'
}

// Convert URL into space delimited string which is required by NaiveBayes model
func normalizeURL(url string) string {
	var urlString string
	url = trimURL(url)
	// get tokens after splitting by slash/dash/dot
	tokens := strings.FieldsFunc(url, split)
	for _, token := range tokens {
		if token != "com" {
			// removing .com since it occurs a lot of times and it should not be
			// included in our features
			urlString = urlString + token + " "
		}
	}
	return urlString
}

// Use Naive Bayes classifier to learn from data/data.csv
func learn() {
	// create the channel of data and errors
	stream := make(chan base.TextDatapoint, 100)
	errors := make(chan error)

	// make a new NaiveBayes model with
	// 2 classes expected (classes in
	// datapoints will now expect {0,1}.
	model = text.NewNaiveBayes(stream, 2, base.OnlyWordsAndNumbers)

	go model.OnlineLearn(errors)

	file, err := os.Open("./data/data.csv")
	checkErr(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), ",")
		if len(s) == 2 {
			url, score := s[0], s[1]
			i := 1
			if score == "bad" {
				i = 0
			}

			stream <- base.TextDatapoint{
				X: normalizeURL(url),
				Y: uint8(i),
			}
		}
	}

	if err = scanner.Err(); err != nil {
		checkErr(err)
	}

	close(stream)
}

// Predict whether a URL is good or bad based on Naive Bayes model,
// class=1/0 1=good, 0=bad
func predict(url string) uint8 {
	nURL := normalizeURL(url)
	class := model.Predict(nURL)
	return class
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
