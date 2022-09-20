package api

import (
	"fmt"
	browser "github.com/EDDYCJY/fake-useragent"
	"net/http"
)

func getWithAgent(url string, args ...interface{}) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(url, args...), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", browser.Random())
	return http.DefaultClient.Do(req)
}
