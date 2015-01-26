package perflog

import (
	"os"
	"time"

	"gopnik"

	"github.com/op/go-logging"
	json "github.com/orofarne/strict-json"
)

type PerfLogEntry struct {
	Coord      gopnik.TileCoord
	Timestamp  time.Time
	RenderTime time.Duration
	SaverTime  time.Duration
}

var log = logging.MustGetLogger("global")
var perflogFile string
var perflogChan chan *PerfLogEntry

func init() {
	perflogChan = make(chan *PerfLogEntry, 1000)
}

func perfLogSaver() {
	f, err := os.OpenFile(perflogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Error("Unable to open stats file: %v", err)
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)

	for entry := range perflogChan {
		err = enc.Encode(entry)
		if err != nil {
			log.Error("PerfLog error: %v", err)
		}
	}
}

func SetupPerflog(fName string) {
	perflogFile = fName
	go perfLogSaver()
}

func SavePerf(perf PerfLogEntry) {
	if perflogFile != "" {
		perflogChan <- &perf
	}
}
