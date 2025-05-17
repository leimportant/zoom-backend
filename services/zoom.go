package services

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

var token string

func GetAccessToken() string {
	if token != "" {
		return token
	}

	clientID := os.Getenv("ZOOM_CLIENT_ID")
	clientSecret := os.Getenv("ZOOM_CLIENT_SECRET")
	accountID := os.Getenv("ZOOM_ACCOUNT_ID")

	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	url := "https://zoom.us/oauth/token?grant_type=account_credentials&account_id=" + accountID

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Authorization", "Basic "+auth)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	var data map[string]interface{}
	json.Unmarshal(body, &data)
	token = data["access_token"].(string)

	return token
}
