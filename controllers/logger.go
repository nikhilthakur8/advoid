package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/nikhilthakur8/advoid/models"
)

func LogDnsQuery(log models.LogDNSQuery) error {
	logTailUrl := "https://s1571226.eu-nbg-2.betterstackdata.com"
	logTailToken := os.Getenv("BETTERSTACK_KEY")
	fmt.Println(logTailToken)
	if logTailToken == "" {
		return nil
	}
	data, err := json.Marshal(log)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", logTailUrl, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+logTailToken)
	if err != nil {
		return err
	}
	_, err = http.DefaultClient.Do(req)
	defer req.Body.Close()
	if err != nil {
		return err
	}
	return nil
}
