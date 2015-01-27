package perflog

import (
	"fmt"
	"os"
	"strings"
	"time"

	json "github.com/orofarne/strict-json"
)

func LoadPerf(fName string, since, until time.Time) ([]PerfLogEntry, error) {
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
			if !since.Equal(time.Time{}) && since.After(entry.Timestamp) {
				continue
			}
			if !until.Equal(time.Time{}) && until.Before(entry.Timestamp) {
				continue
			}
			result = append(result, entry)
		}
	}
	return result, nil
}
