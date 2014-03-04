package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ActiveState/tail"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var (
	config_load_begin = regexp.MustCompile(`\d{4}-\d\d-\d\d \d\d:\d\d:\d\d \+\d\d\d\d \[info\]: using configuration file: <ROOT>`)
	config_load_end   = regexp.MustCompile(`</ROOT>`)
	log_level_pattern = regexp.MustCompile(`\d{4}-\d\d-\d\d \d\d:\d\d:\d\d \+\d\d\d\d \[(.*?)\]: (.*)?`)
)

const (
	INITIALIZE_CONFIG = 0
	ERROR_TYPE        = 1
	WARN_TYPE         = 2
	OTHER_TYPE        = 1000
)

type Setting struct {
	Idobata string `json:idobata`
	Td_Log  string `json:td_log`
}

type Config struct {
	Message []string
	Loading bool
}

func (c Config) messageType(message string) int {
	p := log_level_pattern.FindStringSubmatch(message)

	if p != nil {
		if p[1] == "error" {
			return ERROR_TYPE
		}
		if p[1] == "warn" {
			return WARN_TYPE
		}
	}

	if config_load_begin.FindStringSubmatch(message) != nil {
		return INITIALIZE_CONFIG
	} else {
		return OTHER_TYPE
	}
}

func (c Config) isInitializeLogEnd(message string) bool {
	if config_load_end.FindStringSubmatch(message) != nil {
		c.Loading = false
		return true
	} else {
		return false
	}
}

func (c *Config) appendMessage(message string) {
	c.Message = append(c.Message, message)
}

func main() {
	var c Config

	file, e := ioutil.ReadFile("./setting.json")
	if e != nil {
		os.Exit(1)
	}
	fmt.Printf("%s\n", file)

	var messageType int
	var setting Setting
	json.Unmarshal(file, &setting)

	fmt.Printf("Setting is %v", setting)
	t, err := tail.TailFile(setting.Td_Log, tail.Config{
		Follow: true,
		ReOpen: true})

	if err != nil {
		os.Exit(1)
	}

	for line := range t.Lines {
		messageType = c.messageType(line.Text)

		if messageType == INITIALIZE_CONFIG {
			c.Loading = true
		}

		if messageType == WARN_TYPE ||
			messageType == ERROR_TYPE {
			postMessage(line.Text, setting)
		}

		if c.Loading == true {
			c.appendMessage(line.Text)
		}

		if c.isInitializeLogEnd(line.Text) == true {
			postMessage("[NOW] loading of the td-agent.conf: \n"+strings.Join(c.Message, "\n"), setting)
			c.Message = []string{}
		}

	}
}

func postMessage(message string, setting Setting) {
	data := url.Values{}
	data.Set("source", message)

	client := &http.Client{}

	r, _ := http.NewRequest("POST", setting.Idobata, bytes.NewBufferString(data.Encode()))
	r.Header.Add("User-Agent", "futoase test")

	res, _ := client.Do(r)
	fmt.Println(res.Status)
}
