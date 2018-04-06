package vsts

import (
	"fmt"
	"log"
	"strings"
)

type goTestReview struct {
	diffs       *diffs
	pullRequest *PullRequest
}

func (r *goTestReview) getBotCommentPrefix() string {
	return "[BOT_Image]\n"
}

func (r *goTestReview) review() (bool, error) {
	goSuffix := ".go"
	goTestSuffix := "_test.go"
	commentMsg := fmt.Sprintf("%s\nPlease update test.", r.getBotCommentPrefix())

	var changedGoFiles []string
	var changedGoTestFiles []string

	for _, change := range r.diffs.Changes {
		if strings.HasSuffix(change.Item.Path, goSuffix) {
			changedGoFiles = append(changedGoFiles, change.Item.Path)
		} else if strings.HasSuffix(change.Item.Path, goTestSuffix) {
			changedGoTestFiles = append(changedGoTestFiles, change.Item.Path)
		}
	}

	var missingTestGoFiles []string
	for _, changedGoFile := range changedGoFiles {
		for _, changedGoTestFile := range changedGoTestFiles {
			if strings.EqualFold(strings.TrimSuffix(changedGoTestFile, goTestSuffix), strings.TrimSuffix(changedGoFile, goSuffix)) {
				log.Printf("%s has test update: %s", changedGoFile, changedGoTestFile)
				break
			}
		}
		missingTestGoFiles = append(missingTestGoFiles, changedGoFile)
	}

	commentThreads, err := getCommentThreads(r.pullRequest.Resource.PullRequestID)
	log.Printf("threads: %v", commentThreads.Count)

	if err != nil {
		return false, err
	}

	for _, goFile := range missingTestGoFiles {
		commentThread := commentThread{}
		for _, thread := range commentThreads.Value {
			if !thread.IsDeleted && strings.EqualFold(thread.ThreadContext.FilePath, goFile) {
				for _, comment := range thread.Comments {
					if comment.ID == 1 && comment.Author.ID == config.UserID && strings.HasPrefix(comment.Content, r.getBotCommentPrefix()) {
						commentThread = thread
						break
					}
				}
			}
		}

		if commentThread.Status == "" {
			// create thread
			err := createCommentThread(r.pullRequest.Resource.PullRequestID, goFile, 1, commentMsg)
			if err != nil {
				return false, err
			}
		} else {
			// add comment
			err := addComment(r.pullRequest.Resource.PullRequestID, commentThread, commentMsg, commentMsg)
			if err != nil {
				return false, err
			}
			// set thread active
			err = setCommentThreadStatus(r.pullRequest.Resource.PullRequestID, commentThread, 1)
			if err != nil {
				return false, err
			}
		}
	}

	// review result
	return len(missingTestGoFiles) == 0, nil
}
