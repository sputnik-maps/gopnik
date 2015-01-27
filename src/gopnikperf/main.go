package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"perflog"
)

var perflogFile = flag.String("perflog", "", "Perflog file")
var addr = flag.String("addr", ":8080", "Bind WebUI to addr")
var since = flag.String("since", "", "Use values since time (RFC3339 or @UnixTimestamp)")
var until = flag.String("until", "", "Use values until time (RFC3339 or @UnixTimestamp)")

func parseTime(str string) (time.Time, error) {
	if str == "" {
		return time.Time{}, nil
	}

	if str[0] == '@' {
		var unixts int64
		_, err := fmt.Sscanf(str, "@%v", &unixts)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(unixts, 0), nil
	} else {
		return time.Parse(time.RFC3339, str)
	}
}

func main() {
	flag.Parse()

	sinceT, err := parseTime(*since)
	if err != nil {
		log.Fatalf("Invalid since param: %v", err)
	}
	untilT, err := parseTime(*until)
	if err != nil {
		log.Fatalf("Invalid unitl param: %v", err)
	}

	perfData, err := perflog.LoadPerf(*perflogFile, sinceT, untilT)
	if err != nil {
		log.Fatalf("Failed to load performance log: %v", err)
	}

	log.Printf("%v items loaded", len(perfData))

	runWebUI(*addr, perfData)
}
