package ext

import (
	"fmt"
	"net/http"
)

// GetHeaderFromHealthCheck gets header from health check request
func GetHeaderFromHealthCheck(done chan<- http.Header, endpoint string) error {
	urlString := fmt.Sprintf("%s/%s", endpoint, "healthcheck")
	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		done <- nil
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		done <- nil
		return err
	}
	done <- resp.Header
	return nil
}
