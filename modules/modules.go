package modules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"io"
	"net/http"
	"time"
)

// Options sets individual reporting options
type Options struct {
	Method           string
	URL              string
	interval         int
	AcceptableStatus []int `yaml:"acceptableStatus"`
	Min              int
	Timeout          int
	Body             string
	Headers          map[string]string
}

// Reporter provides the structure for each reporter
type Reporter struct {
	Name     string
	Module   string
	Interval int
	Options
}

// Status is the returned message from a success or error event
type Status struct {
	Message string
	Error   string
	Name    string
	Code    string
}

// tile38ResponseStats is the stats returned from the tile38Response
type tile38ResponseStats struct {
	NumObjects int `json:"num_objects"`
}

// tile38Response is the network response from Tile38
type tile38Response struct {
	Stats tile38ResponseStats `json:"stats"`
}

// HTTP module allows for generic http GET/POST requests to endpoints
func HTTP(reporter Reporter) (Status, error) {
	var response *http.Response
	var err error
	timeout := time.Duration(30 * time.Second)
	if reporter.Options.Timeout != 0 {
		timeout = time.Duration(reporter.Options.Timeout) * time.Second
	}
	client := http.Client{Timeout: timeout}
	req, _ := http.NewRequest("GET", reporter.Options.URL, nil)

	if reporter.Method == "POST" {
		req, _ = http.NewRequest("POST", reporter.Options.URL, bytes.NewBuffer([]byte(reporter.Options.Body)))
		req.Header.Add("Content-Type", "application/json")
	}
	req.Header.Add("Cache-Control", "no-cache")

	for key, value := range reporter.Options.Headers {
		req.Header.Add(key, value)
	}

	response, err = client.Do(req)
	if err != nil {
		errorStatus := Status{
			Message: "error",
			Error:   fmt.Sprintf("%s", err),
		}
		if response != nil {
			errorStatus.Code = string(response.StatusCode)
		}
		return errorStatus, err
	}
	defer response.Body.Close()
	status := Status{
		Name: reporter.Name,
		Code: string(response.StatusCode),
	}
	for _, statusCode := range reporter.Options.AcceptableStatus {
		if statusCode == response.StatusCode {
			status.Message = "success"
			return status, nil
		}
	}
	status.Message = "error"
	status.Error = fmt.Sprintf("Wanted status: %d got: %d", reporter.Options.AcceptableStatus, response.StatusCode)
	return status, nil
}

// Tile38 module allows for checking the object_count in a tile38 http instance
func Tile38(reporter Reporter) (Status, error) {
	url := fmt.Sprintf("%s/server", reporter.Options.URL)

	timeout := time.Duration(30 * time.Second)
	if reporter.Options.Timeout != 0 {
		timeout = time.Duration(reporter.Options.Timeout) * time.Second
	}
	client := http.Client{Timeout: timeout}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(reporter.Options.Body)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")

	response, err := client.Do(req)

	if err != nil {
		errorStatus := Status{
			Message: "error",
			Error:   fmt.Sprintf("%s", err),
		}
		if response != nil {
			errorStatus.Code = string(response.StatusCode)
		}

		return errorStatus, err
	}
	defer response.Body.Close()
	status := Status{
		Name: reporter.Name,
		Code: string(response.StatusCode),
	}

	buf := make([]byte, response.ContentLength)
	if _, err = io.ReadFull(response.Body, buf); err != nil {
		errorStatus := Status{
			Message: "error",
			Error:   fmt.Sprintf("%s", err),
		}
		return errorStatus, err
	}
	var tile38Body tile38Response
	err = json.Unmarshal([]byte(buf), &tile38Body)
	if tile38Body.Stats.NumObjects >= reporter.Options.Min {
		status.Message = "success"
		return status, nil
	}

	status.Message = "error"
	status.Error = fmt.Sprintf("Objects below minimum threshold: %d got: %d", reporter.Options.Min, tile38Body.Stats.NumObjects)
	return status, nil
}

// Redis module allows for checking if a redis instance is online
func Redis(reporter Reporter) (Status, error) {
	opts, err := redis.ParseURL(reporter.URL)
	if err != nil {
		errorStatus := Status{
			Message: "error",
			Error:   fmt.Sprintf("invalid URL %s %e", reporter.URL, err),
		}
		return errorStatus, nil
	}
	if reporter.Timeout != 0 {
		timeout := time.Duration(reporter.Timeout) * time.Second
		opts.DialTimeout = timeout
		opts.ReadTimeout = timeout
		opts.WriteTimeout = timeout
	}
	client := redis.NewClient(opts)
	pong, err := client.Ping().Result()
	client.Close()
	if err != nil {
		errorStatus := Status{
			Message: "error",
			Error:   fmt.Sprintf("%s %e", reporter.URL, err),
		}
		return errorStatus, err
	}
	if pong != "PONG" {
		errorStatus := Status{
			Message: "error",
			Error:   "Unable to ping redis -- did not get a response",
		}
		return errorStatus, nil
	}
	status := Status{
		Message: "success",
	}
	return status, nil
}
