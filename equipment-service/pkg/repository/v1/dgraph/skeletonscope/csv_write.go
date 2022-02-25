package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
)

func main() {
	// appendCoreFactor("equipment_server.csv")
	correctSAG("equipment_server.csv")
}

func appendCoreFactor(filePath string) {

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("", err)
	}
	defer f.Close()
	r := csv.NewReader(bufio.NewReader(f))
	r.Comma = ';'
	file, err := os.Create("equipment_server_corefactor.csv")
	if err != nil {
		log.Fatal("", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	w.Comma = ';'
	defer w.Flush()

	header, err := r.Read()
	header = append(header, "corefactor_oracle")
	w.Write(header)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		coreFactor := float32(rand.Intn(4)+1) * 0.25
		s := fmt.Sprintf("%f", coreFactor)
		// fmt.Println(s)
		// record = record[:len(record)-1]
		record = append(record, s)
		w.Write(record)
	}
}

func correctSAG(filePath string) {

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("", err)
	}
	defer f.Close()
	r := csv.NewReader(bufio.NewReader(f))
	r.Comma = ';'
	file, err := os.Create("equipment_server_sag.csv")
	if err != nil {
		log.Fatal("", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	w.Comma = ';'
	defer w.Flush()

	header, err := r.Read()
	w.Write(header)
	index := -1
	for i := range header {
		if header[i] == "sag" {
			index = i
			break
		}
	}
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if len(record) < index {
			continue
		}

		if record[index] == "C" {
			record[index] = "1.2"
		}

		record[index] = strings.Replace(record[index], ",", ".", -1)

		w.Write(record)
	}
}
