package utils

import (
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

func MakeAPIRequest(reqURI string, header map[string]string, method string, data io.Reader) ([]byte, error) {
	// Create HTTP client and request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create HTTP request
	req, err := http.NewRequest(method, reqURI, data)

	// Check for errors in creating the request
	if err != nil {
		return nil, err
	}

	// Set headers for the request
	for key, value := range header {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)

	// Check for errors in making the request
	if err != nil {
		return nil, err
	}

	// Ensure the response body is closed after reading
	defer resp.Body.Close()

	// Check for non-200 status codes
	if resp.StatusCode != 200 {
		return nil, errors.New("API request failed with status code: " + resp.Status)
	}

	// Read the response body in bytes
	respData, err := io.ReadAll(resp.Body)

	// Check for errors in reading the response body
	if err != nil {
		log.Println("Error reading response body:", err)
		return nil, err
	}

	// Return the response data and no error
	return respData, nil
}
