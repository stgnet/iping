package main

import (
	"fmt"
	"github.com/stgnet/iping"
)

func main() {
	options := iping.Options{Target: "8.8.8.8", Count: 3}
	results, err := options.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", results)
}
