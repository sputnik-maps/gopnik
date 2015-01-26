package main

import (
	"flag"
	"log"
	"perflog"
)

var perflogFile = flag.String("perflog", "", "Perflog file")
var addr = flag.String("addr", ":8080", "Bind WebUI to addr")

func main() {
	flag.Parse()

	perfData, err := perflog.LoadPerf(*perflogFile)
	if err != nil {
		log.Fatalf("Failed to load performance log: %v", err)
	}

	log.Printf("%v items loaded", len(perfData))

	runWebUI(*addr, perfData)
}
