package cron

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// GetAccessToken returns the cached access token
func GetAccessToken() string {
	tokenMutex.RLock()
	defer tokenMutex.RUnlock()
	return accessToken
}

// StartTokenRefresher starts the token refresh routine
func StartTokenRefresher(corpID string, secret string) {
	refreshToken(corpID, secret)
}

func refreshToken(corpID string, secret string) {
	url := fmt.Sprintf("%s?corpid=%s&corpsecret=%s", WxTokenURL, corpID, secret)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to get access token: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return
	}

	var tokenResp AccessTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		log.Printf("Failed to unmarshal response: %v", err)
		return
	}

	if tokenResp.ErrCode != 0 {
		log.Printf("Error getting access token: %s", tokenResp.ErrMsg)
		return
	}

	tokenMutex.Lock()
	accessToken = tokenResp.AccessToken
	tokenMutex.Unlock()

	// Schedule next refresh 10 seconds before expiration
	refreshTime := time.Duration(tokenResp.ExpiresIn-10) * time.Second
	if timer != nil {
		timer.Stop()
	}
	timer = time.AfterFunc(refreshTime, func() {
		refreshToken(corpID, secret)
	})

	log.Printf("Access token refreshed, will refresh again in %v", refreshTime)
}
