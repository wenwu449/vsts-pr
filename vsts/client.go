package vsts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func getFromVsts(url string, v interface{}) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(config.Username, config.Password)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("repsonse with non 200 code of %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

func postToVsts(url string, v interface{}) error {
	return sendToVsts("POST", url, v)
}

func putToVsts(url string, v interface{}) error {
	return sendToVsts("PUT", url, v)
}

func patchToVsts(url string, v interface{}) error {
	return sendToVsts("PATCH", url, v)
}

func sendToVsts(method string, url string, v interface{}) error {
	client := &http.Client{}
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(v)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	req.SetBasicAuth(config.Username, config.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println(resp.Status)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("repsonse with non 200|201 code of %d", resp.StatusCode)
	}

	return nil
}
