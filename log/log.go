package log

import "fmt"

const Debug = true
const Verbose = false

// Error messages
func Error(msg ...interface{}) {
	fmt.Println(msg)
}

// Warning messages
func Warning(msg ...interface{}) {
	fmt.Println(msg)
}

// Info is less important
func Info(msg ...interface{}) {
	if Debug {
		fmt.Println(msg)
	}
}

// Bullshit that you usually don't want to hear
func Bullshit(msg ...interface{}) {
	if Debug && Verbose {
		fmt.Println(msg)
	}
}
