package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/wenwu449/vsts-pr/vsts"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config, err := vsts.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

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
}
