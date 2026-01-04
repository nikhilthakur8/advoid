package logger

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nikhilthakur8/advoid/definitions"
)

var LogChan = make(chan definitions.DNSLog, 10_000)

func EmitLogs(logs definitions.DNSLog) {
	select {
	case LogChan <- logs:
	default:
		// drop logs if channel is full
	}
}

func startSideCarSender(sidecarURL string) {
	client := http.Client{
		Timeout: 200 * time.Millisecond,
	}

	for log := range LogChan {
		data, err := json.Marshal(log)
		if err != nil {
			continue
		}

		req, err := http.NewRequest("POST", sidecarURL, bytes.NewBuffer(data))
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
	}

}

func init() {
	go startSideCarSender("http://localhost:9000/ingest")
}
