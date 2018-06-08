package main

import (
	"encoding/csv"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

var (
	timiUserWhiteList = make(map[string]string)
)

func readUserFile(file string) {
	in, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("read user file: %v failed, error: %v\n", file, err)
	}
	r := csv.NewReader(strings.NewReader(string(in)))

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("parse user file: %v failed, error: %v\n", file, err)
		}

		if len(record) != 0 {
			timiUserWhiteList[record[1]] = record[2]
		}
	}
}

func isIllegalUser(chineseName, englishName string) bool {
	if name, isPresent := timiUserWhiteList[chineseName]; isPresent {
		if name == englishName {
			return true
		}
	}
	return false
}
