package queue

import (
	"bufio"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/knutaldrin/elevator/log"
)

const filename string = "OrderLog.txt"

// ReadLog reads the log and returns an int slice of floors
func ReadLog() []int {
	file, err := os.Open(filename)
	if err != nil {
		log.Info(err)
		ioutil.WriteFile(filename, []byte(""), 0666)
		file, _ = os.Open(filename)
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	var nSlice []int

	for {
		s, err := reader.ReadString('\n')
		n, sErr := strconv.ParseInt(strings.Replace(strings.Replace(s, "\r", "", -1), "\n", "", -1), 10, 32) // removes \r\n and parses string
		if err != nil {
			break
		}
		log.Check(err)
		log.Check(sErr)
		nSlice = append(nSlice, int(n))
	}

	return nSlice
}

//writeLog formats its int slice argument and overwrites the log with it
func writeLog(ns []int) {
	var resultSlice []byte
	var appendSlice []byte

	for i := 0; i < len(ns); i++ {
		appendSlice = []byte(strconv.FormatInt(int64(ns[i]), 10) + "\n")
		for j := 0; j < len(appendSlice); j++ {
			resultSlice = append(resultSlice, appendSlice[j])
		}
	}

	ioutil.WriteFile(filename, resultSlice, 0666)
}

//isInLog is a boolean check of whether a floor is recorded in the log.
func isInLog(floor int) bool {
	intSlice := ReadLog()
	for i := 0; i < len(intSlice); i++ {
		if floor == intSlice[i] {
			return true
		}
	}
	return false
}

//RemoveFromLog removes a floor from the log file, shortening the file by one character. If the floor is not present in the log, nothing happens.
func RemoveFromLog(floor int) {
	oldSlice := ReadLog()
	var newSlice []int

	for i := 0; i < len(oldSlice); i++ {
		if oldSlice[i] != floor {
			newSlice = append(newSlice, oldSlice[i])
		}
	}
	writeLog(newSlice)

	log.Debug("Removed from log: ", floor)
}

//AddToLog adds a floor from the log file, if the floor is not already in the queue. If added, the file size increases by one character.
func AddToLog(floor int) {
	if isInLog(floor) {
		return
	}

	intSlice := append(ReadLog(), floor)

	writeLog(intSlice)

	log.Debug("Logged floor: ", floor)
}
