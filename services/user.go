package services

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/nikhilthakur8/advoid/config"
	"github.com/nikhilthakur8/advoid/definitions"
	"github.com/nikhilthakur8/advoid/utils"
)

type BodyData struct {
	Data definitions.UserConfig `json:"data"`
}

func MakeAPIRequest(reqURI string, header map[string]string, method string) ([]byte, error) {
	// Create HTTP client and request
	client := &http.Client{}

	// Create HTTP request
	req, err := http.NewRequest(method, reqURI, nil)

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

func getUserConfigFromServer(userId string) (definitions.UserConfig, error) {

	// Make a HTTP Client Request to fetch user config
	reqURI := config.GetEnv("BACKEND_URI") + "/admin/user-configs/" + userId

	header := map[string]string{
		"Authorization": "Bearer " + config.GetEnv("API_KEY"),
	}

	// Make the API request to fetch user config
	resBytes, err := MakeAPIRequest(reqURI, header, "GET")

	// Check for errors in making the API request
	if err != nil {
		log.Println("Error making API request for UserConfig:", err)
		// Return an empty UserConfig and the error
		return definitions.UserConfig{}, err
	}

	respData := BodyData{}

	// Unmarshal the response bytes into the BodyData struct
	err = json.Unmarshal(resBytes, &respData)

	if err != nil {
		log.Println("Error unmarshalling JSON for UserConfig:", err)
		// Return an empty UserConfig and the error
		return definitions.UserConfig{}, err
	}

	userConfig := respData.Data
	// Return the userConfig and no error
	return userConfig, nil
}

func GetUserConfigFromCache(userId string) (definitions.UserConfig, error) {
	// Get cache instance
	cache := utils.GetCacheInstance()

	key := "USER_CONFIG_" + userId

	// Try to get userConfig from cache
	if val, found := cache.Get(key); found {
		return val.(definitions.UserConfig), nil
	}

	// If not found in cache, fetch from server
	userConfig, err := getUserConfigFromServer(userId)

	if err != nil {
		return definitions.UserConfig{}, err
	}

	// Set the fetched userConfig in cache with a TTL of 1 minute
	cache.Set(key, userConfig, 1*time.Minute)

	return userConfig, nil
}
