package main

import (
	"fmt"
)

type VerboseLogger bool

func (vl VerboseLogger) Print(args ...interface{}) {
	if vl == true {
		fmt.Print(args...)
	}
}

func (vl VerboseLogger) Println(args ...interface{}) {
	if vl == true {
		fmt.Println(args...)
	}
}
func (vl VerboseLogger) Printf(format string, args ...interface{}) {
	if vl == true {
		fmt.Printf(format, args...)
	}
}
