package vsts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Client struct {
	host     string
	username string
	password string
}

func NewClient(host, username, password string) *Client {
	return &Client{
		host:     host,
		username: username,
		password: password,
	}
}

func (c *Client) getFromVsts(url string, v interface{}) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.username, c.password)
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

func (c *Client) postToVsts(url string, v interface{}) error {
	return c.sendToVsts("POST", url, v)
}

func (c *Client) putToVsts(url string, v interface{}) error {
	return c.sendToVsts("PUT", url, v)
}

func (c *Client) patchToVsts(url string, v interface{}) error {
	return c.sendToVsts("PATCH", url, v)
}

func (c *Client) sendToVsts(method string, url string, v interface{}) error {
	client := &http.Client{}
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(v)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.username, c.password)
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
