package services

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nikhilthakur8/advoid/config"
	"github.com/nikhilthakur8/advoid/definitions"
	"github.com/nikhilthakur8/advoid/utils"
)

type BodyData struct {
	Data definitions.UserConfig `json:"data"`
}

func GetUserConfigFromServer(userId string) (definitions.UserConfig, error) {

	// Make a HTTP Client Request to fetch user config
	reqURI := config.GetEnv("BACKEND_URI") + "/admin/user-configs/" + userId

	header := map[string]string{
		"Authorization": "Bearer " + config.GetEnv("API_KEY"),
	}

	// Make the API request to fetch user config
	resBytes, err := utils.MakeAPIRequest(reqURI, header, "GET")

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
	userConfig, err := GetUserConfigFromServer(userId)

	if err != nil {
		return definitions.UserConfig{}, err
	}

	// Set the fetched userConfig in cache with a TTL of 1 minute
	cache.Set(key, userConfig, 1*time.Minute)

	return userConfig, nil
}
