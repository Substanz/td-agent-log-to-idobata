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

type Setting struct {
	Idobata string `json:idobata`
	Td_Log  string `json:td_log`
}

func main() {

	file, e := ioutil.ReadFile("./setting.json")
	if e != nil {
		os.Exit(1)
	}
	fmt.Printf("%s\n", file)

	var setting Setting
	json.Unmarshal(file, &setting)

	fmt.Printf("Setting is %v", setting)
	t, err := tail.TailFile(setting.Td_Log, tail.Config{
		Follow: true,
		ReOpen: true})

	if err != nil {
		os.Exit(1)
	}

	firstMatch, _ := regexp.Compile(`\d{4}-\d\d-\d\d \d\d:\d\d:\d\d \+\d\d\d\d \[info\]: using configuration file: <ROOT>`)
	lastMatch, _ := regexp.Compile(`</ROOT>`)

	var firstResult []string
	var lastResult []string
	var configMessage []string
	var flag bool

	for line := range t.Lines {
		firstResult = firstMatch.FindStringSubmatch(line.Text)
		lastResult = lastMatch.FindStringSubmatch(line.Text)

		if firstResult != nil {
			flag = true
			continue
		}

		if lastResult != nil {
			flag = false
			postMessage("td-agentがconcatしたtd-agentの設定です: \n"+strings.Join(configMessage, "\n"), setting)
			configMessage = []string{}
		}

		if flag == true {
			configMessage = append(configMessage, line.Text)
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
