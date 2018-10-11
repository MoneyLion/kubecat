package main

import (
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/stevelacy/kubecat/modules"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// Config for kubecat
type Config struct {
	Reporters []modules.Reporter
}

var config Config

func main() {
	raw, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		panic("unable to read ./config.yaml")
	}
	parsed := string(raw)

	client, err := raven.New(os.Getenv("SENTRY_DSN"))
	if err != nil {
		panic("unable to connect to sentry")
	}

	err = yaml.Unmarshal([]byte(parsed), &config)
	if err != nil {
		panic("unable to parse config yaml")
	}

	for _, reporter := range config.Reporters {
		fmt.Printf("Loading reporter: %s with module: %s\n", reporter.Name, reporter.Module)
		if strings.Contains(reporter.Options.URL, "env:") {
			env := strings.Replace(reporter.Options.URL, "env:", "", 1)
			reporter.Options.URL = os.Getenv(env)
		}
		runModule(reporter)
		startChannel(reporter, client)
	}

	go forever()
	select {}
}

func startChannel(reporter modules.Reporter, client *raven.Client) {
	ticker := time.NewTicker(time.Duration(reporter.Interval) * time.Second)
	go func() {
		for _ = range ticker.C {
			status := runModule(reporter)
			if status.Error != "" {
				notifySentry(client, reporter, status)
			}
		}
	}()
}

func runModule(reporter modules.Reporter) modules.Status {
	if reporter.Module == "http" {
		status, _ := modules.HTTP(reporter)
		return status
	}
	if reporter.Module == "Tile38" {
		status, _ := modules.Tile38(reporter)
		return status
	}
	if reporter.Module == "Redis" {
		status, _ := modules.Redis(reporter)
		return status
	}
	fmt.Printf("No module found for reporter '%s' -- %s\n", reporter.Name, reporter.Module)
	return modules.Status{}
}

func notifySentry(client *raven.Client, reporter modules.Reporter, status modules.Status) {
	fmt.Printf("reporting: %s, error: %s \n", reporter.Name, status.Error)
	message := fmt.Sprintf("%s reporter: %s", status.Message, reporter.Name)
	title := fmt.Sprintf("Error: %s is unreachable \n %s", reporter.Name, status.Error)
	messages := map[string]string{
		"message": message,
		"url":     reporter.Options.URL,
	}
	raven.CaptureMessage(title, messages)
}

func forever() {
	for {
		time.Sleep(time.Second)
	}
}
