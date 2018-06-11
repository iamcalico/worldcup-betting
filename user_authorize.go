package main

import (
	"encoding/csv"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// TODO: 更为优雅的做法是抽象出一个公共函数
func readUserFile(original string, newUser string) {
	f, err := os.OpenFile(config.TimiNewUser, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("create file %v failed, reason: %v\n", config.TimiNewUser, err)
	}
	timiNewUsers = f

	originalIn, err := ioutil.ReadFile(original)
	if err != nil {
		log.Fatalf("read user file: %v failed, error: %v\n", original, err)
	}
	originalRead := csv.NewReader(strings.NewReader(string(originalIn)))

	newUserIn, err := ioutil.ReadFile(newUser)
	if err != nil {
		log.Fatalf("read user file: %v failed, error: %v\n", newUserIn, err)
	}
	newUserRead := csv.NewReader(strings.NewReader(string(newUserIn)))

	for {
		record, err := originalRead.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("parse user file: %v failed, error: %v\n", original, err)
		}

		if len(record) != 0 {
			timiUserWhiteList[record[1]] = record[2]
		}
	}

	for {
		record, err := newUserRead.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("parse user file: %v failed, error: %v\n", newUser, err)
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

func addNewUser(file *os.File, chineseName, englishName string) bool {
	timiUserWhiteList[chineseName] = englishName
	record := []string{"X", chineseName, englishName}
	w := csv.NewWriter(file)
	if err := w.Write(record); err != nil {
		handleError(err)
		return false
	}
	w.Flush()
	return true
}
