package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	apiKey := strings.Fields(headers.Get("Authorization"))

	if len(apiKey) < 2 || apiKey[1] == "" {
		return "", fmt.Errorf("Api Key not found")
	}
	return apiKey[1], nil

}
