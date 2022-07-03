package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/acarl005/stripansi"
)

var wg sync.WaitGroup

func main() {
	var oneLine, verboseMode, plainTextMode bool
	var webhookURL, lines string
	flag.StringVar(&webhookURL, "u", "", "Slack Webhook URL")
	flag.BoolVar(&oneLine, "1", false, "Send message line-by-line")
	flag.BoolVar(&verboseMode, "v", false, "Verbose mode")
	flag.BoolVar(&plainTextMode, "p", false, "Plain text mode")
	flag.Parse()

	webhookEnv := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookEnv != "" {
		webhookURL = webhookEnv
	}

	if webhookURL == "" {
		fmt.Fprintln(os.Stderr, "Slack Webhook URL not set!")
		os.Exit(1)
	}

	if !isStdin() {
		os.Exit(1)
	}

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		line := sc.Text()

		fmt.Println(line)
		if oneLine {
			if !plainTextMode {
				line = codeBlock(line)
			}
			wg.Add(1)
			go slackCat(webhookURL, line)
		} else {
			lines += line
			lines += "\n"
		}
	}

	if !oneLine {
		if !plainTextMode {
			lines = codeBlock(lines)
		}
		wg.Add(1)
		go slackCat(webhookURL, lines)
	}
	wg.Wait()
}

func isStdin() bool {
	f, e := os.Stdin.Stat()
	if e != nil {
		return false
	}

	if f.Mode()&os.ModeNamedPipe == 0 {
		return false
	}

	return true
}

type data struct {
	Text string `json:"text"`
}

func slackCat(url string, line string) {
	data, _ := json.Marshal(data{Text: stripansi.Strip(line)})
	http.Post(url, "application/json", strings.NewReader(string(data)))
	wg.Done()
}

func codeBlock(s string) string {
	return fmt.Sprintf("```%s```", s)
}
