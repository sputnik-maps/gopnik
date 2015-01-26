package perflog

import (
	"fmt"
	"os"
	"strings"

	json "github.com/orofarne/strict-json"
)

func LoadPerf(fName string) ([]PerfLogEntry, error) {
	fIn, err := os.Open(fName)
	if err != nil {
		return nil, err
	}
	defer fIn.Close()

	dec := json.NewDecoder(fIn)
	var result []PerfLogEntry
	for {
		var entry PerfLogEntry
		err = dec.Decode(&entry)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				break
			} else {
				return nil, fmt.Errorf("Decode error: %v", err)
			}
		} else {
			result = append(result, entry)
		}
	}
	return result, nil
}
