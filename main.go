package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/wenwu449/vsts-pr/vsts"
)

const (
	logPath = "LOG_PATH"
)

func main() {
	fmt.Println("a")

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	logFilePath := os.Getenv(logPath)
	if len(logFilePath) > 0 {

		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			log.Fatal(err)
		}

		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)
	}
	log.Println("log: %s", logFilePath)

	cmd := flag.String("command", "", "The command")
	flag.Parse()
	fmt.Printf("cmd: %v", *cmd)

	if !strings.EqualFold(*cmd, "pr") {
		fmt.Println("cmd != 'pr', exit")
		return
	}

	config := &vsts.Config{}
	var err error = nil
	//config, err := vsts.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	switch *cmd {
	case "pr":
		pr, err := vsts.ParsePullRequest()
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Got PR update: %v\n", pr.Resource.PullRequestID)

		if !strings.EqualFold(pr.Resource.TargetRefName, fmt.Sprintf("%s/%s", "refs/heads", config.MasterBranch)) {
			fmt.Printf("unexpected target branch: %s\n", pr.Resource.TargetRefName)
			return
		}

		err = vsts.Review(pr)
		if err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Printf("unknown command: '%s'", *cmd)
	}
}
