package log

import "fmt"
import cli "github.com/ivpusic/go-clicolor/clicolor"

const debug = true
const verbose = true

// Error messages
func Error(msg ...interface{}) {
	cli.Print("ERROR: " + fmt.Sprint(msg...)).In("red")
}

// Check an error, if != then log Error
func Check(err error) {
	if err != nil {
		Error(err)
	}
}

// Warning messages
func Warning(msg ...interface{}) {
	cli.Print("Warning: " + fmt.Sprint(msg...)).In("yellow")
}

// Info is usually good
func Info(msg ...interface{}) {
	cli.Print(fmt.Sprint(msg...)).In("green")
}

// Text is normal text
func Text(msg ...interface{}) {
	fmt.Println(msg)
}

// Debug messages
func Debug(msg ...interface{}) {
	if debug {
		cli.Print(fmt.Sprint(msg...)).In("blue")
	}
}

// Bullshit that you usually don't want to hear
func Bullshit(msg ...interface{}) {
	if debug && verbose {
		cli.Print(fmt.Sprint(msg...)).In("magenta")
	}
}
