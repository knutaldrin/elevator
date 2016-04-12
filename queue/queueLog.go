package queue

import (
	"bytes"
	"io/ioutil"

	"github.com/knutaldrin/elevator/log"
)

//bruker tekstfil/UTF-8 for enkel debugging

const filename string = "OrderLog.txt"
const nFloors int = 4

func readLog() string { //tested
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Info(err)
		ioutil.WriteFile(filename, []byte(""), 0666)
		dat, _ = ioutil.ReadFile(filename)
	}

	return string(dat)
}

func writeLog(s string) { //tested
	err := ioutil.WriteFile(filename, []byte(s), 0666)
	if err != nil {
		log.Info(err)
	}
}

func stringToIntSlice(s string) []int { //tested
	var nSlice []int

	for _, rn := range s { //Iterates over runes
		nSlice = append(nSlice, int(rn)-'0')
	}

	return nSlice
}

func intSliceToString(ns []int) string { //tested
	var buffer bytes.Buffer

	for i := 0; i < len(ns); i++ {
		buffer.WriteString(string(ns[i] + '0'))
	}

	return buffer.String()
}

func IsInLog(floor int) bool { //tested
	intSlice := stringToIntSlice(readLog())
	for i := 0; i < len(intSlice); i++ {
		if floor == intSlice[i] {
			return true
		}
	}
	return false
}

func RemoveFromLog(floor int) { //tested
	oldSlice := stringToIntSlice(readLog())
	var newSlice []int

	for i := 0; i < len(oldSlice); i++ {
		if oldSlice[i] != floor {
			newSlice = append(newSlice, oldSlice[i])
		}
	}
	writeLog(intSliceToString(newSlice))
}

func AppendToLog(floor int) { //tested
	if isInLog(floor) {
		return
	}

	intSlice := stringToIntSlice(readLog())

	intSlice = append(intSlice, floor)

	writeLog(intSliceToString(intSlice))
}

//nextInLog?
