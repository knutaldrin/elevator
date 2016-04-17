package queue

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/knutaldrin/elevator/log"
)

//bruker tekstfil/UTF-8 for enkel debugging

const filename string = "OrderLog.txt"

//readFile leser loggen og returnerer innholdet som en int slice
func readLog() []int {
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
		n, sErr := strconv.ParseInt(strings.Replace(strings.Replace(s, "\r", "", -1), "\n", "", -1), 10, 32) //Fjerner \r\n og parser streng
		if err == io.EOF {
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

//IsInLog is a boolean check of whether a floor is recorded in the log.
func IsInLog(floor int) bool {
	intSlice := readLog()
	for i := 0; i < len(intSlice); i++ {
		if floor == intSlice[i] {
			return true
		}
	}
	return false
}

//RemoveFromLog removes a floor from the log file, shortening the file by one character. If the floor is not present in the log, nothing happens.
func RemoveFromLog(floor int) {
	oldSlice := readLog()
	var newSlice []int

	for i := 0; i < len(oldSlice); i++ {
		if oldSlice[i] != floor {
			newSlice = append(newSlice, oldSlice[i])
		}
	}
	writeLog(newSlice)
}

//AppendToLog adds a floor from the log file, if the floor is not already in the queue. If added, the file size increases by one character.
func AppendToLog(floor int) {
	if IsInLog(floor) {
		return
	}

	intSlice := append(readLog(), floor)

	writeLog(intSlice)
}

//nextInLog?

/*func main() {
	fmt.Print(readLog(), "\n")
	writeLog([]int{6, 1, 3, 1231, 87, 132, 55555555})
	fmt.Print("\n")
	fmt.Print(readLog(), "\n")
	AppendToLog(55)
	fmt.Print(readLog(), "\n")
	RemoveFromLog(55555555)
	fmt.Print(readLog(), "\n")
}*/
