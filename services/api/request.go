package api

import (
	"fmt"
	"net/http"
)

func getWithAgent(url string, args ...interface{}) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(url, args...), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Origin", "https://live.bilibili.com")
	req.Header.Set("Referer", "https://live.bilibili.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	return http.DefaultClient.Do(req)
}
