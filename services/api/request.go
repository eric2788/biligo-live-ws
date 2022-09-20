package api

import (
	"fmt"
	"github.com/corpix/uarand"
	"net/http"
)

func getWithAgent(url string, args ...interface{}) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(url, args...), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", uarand.GetRandom())
	return http.DefaultClient.Do(req)
}
