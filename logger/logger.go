package logger

import (
	"bytes"
	"encoding/json"
	"sync"
	"time"

	"github.com/nikhilthakur8/advoid/config"
	"github.com/nikhilthakur8/advoid/definitions"
	"github.com/nikhilthakur8/advoid/utils"
)

var (
	logBuffer   []definitions.DNSLog
	bufferMutex sync.Mutex
	BATCH_SIZE  = 10_000
)

func AddLogEntry(entry definitions.DNSLog) {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()
	logBuffer = append(logBuffer, entry)

	shouldFlush := len(logBuffer) >= BATCH_SIZE
	if shouldFlush {
		go flushLogs()
	}
}

func flushLogs() {
	bufferMutex.Lock()
	if len(logBuffer) == 0 {
		return
	}

	batch := logBuffer
	logBuffer = []definitions.DNSLog{}
	bufferMutex.Unlock()

	payload, _ := json.Marshal(map[string]interface{}{
		"logs": batch,
	})

	reqURI := config.GetEnv("BACKEND_URI") + "/admin/logs/batch"
	header := map[string]string{
		"Authorization": "Bearer " + config.GetEnv("API_KEY"),
	}
	_, err := utils.MakeAPIRequest(reqURI, header, "POST", bytes.NewBuffer(payload))
	if err != nil {
		// If there's an error, re-add the logs back to the buffer
		bufferMutex.Lock()
		logBuffer = append(logBuffer, batch...)
		bufferMutex.Unlock()
		return
	}
}

func StartAutoFlush() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			flushLogs()
		}
	}()
}
