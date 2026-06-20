package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var (
	host string
	pat  string
)

func api(method, path string, body, out any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	url := strings.TrimRight(host, "/") + "/api/" + strings.TrimLeft(path, "/")
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+pat)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		errBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("%s: %s", res.Status, strings.TrimSpace(string(errBody)))
	}

	if out != nil && res.StatusCode != 204 {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			return fmt.Errorf("decode: %w", err)
		}
	}
	return nil
}

func get(path string, out any) error {
	return api("GET", path, nil, out)
}

func post(path string, body, out any) error {
	return api("POST", path, body, out)
}

func put(path string, body, out any) error {
	return api("PUT", path, body, out)
}
