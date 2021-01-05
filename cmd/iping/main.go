package main

import (
	"fmt"
	"github.com/stgnet/iping"
)

func main() {
	options := iping.Options{Target: "8.8.8.8", Count: 3}
	results, err := options.Ping()
	if err != nil {
		fmt.Printf("Ping failed: %v\n", err)
	} else {
		// fmt.Printf("%#v\n", results)
		fmt.Printf("        IP: %s\n", results.IP.String())
		fmt.Printf("      Sent: %d\n", results.Sent)
		fmt.Printf("     Recvd: %d\n", results.Received)
		fmt.Printf("  Response: %#v\n", results.Response)
		fmt.Printf("ResponseMs: %#v\n", results.ResponseMs)
		fmt.Printf("   Average: %v\n", results.Average)
		fmt.Printf("Average MS: %v\n", results.Average.Milliseconds())
		fmt.Printf("AverageSec: %v\n", results.Average.Seconds())
	}
}
