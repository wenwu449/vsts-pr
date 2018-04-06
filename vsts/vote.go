package vsts

import (
	"log"
	"strconv"
	"strings"
)

func getReviewerURL(pullRequestID int) string {
	reviewerURLTemplate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/pullRequests/{pullRequest}/reviewers/{reviewer}?api-version={version}"
	r := strings.NewReplacer(
		"{instance}", config.Instance,
		"{project}", config.Project,
		"{repository}", config.Repo,
		"{pullRequest}", strconv.Itoa(pullRequestID),
		"{reviewer}", config.UserID,
		"{version}", "3.0-preview")

	return r.Replace(reviewerURLTemplate)
}

func votePullRequest(pullRequestID int, vote int) error {
	log.Printf("Vote on PR %v: %v...\n", pullRequestID, vote)

	putVote := putVote{
		Vote: vote,
	}

	url := getReviewerURL(pullRequestID)

	err := putToVsts(url, putVote)
	if err != nil {
		return err
	}

	return nil
}
