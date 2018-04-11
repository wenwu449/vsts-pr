package vsts

import (
	"log"
	"strings"
)

var config *Config

func init() {
	var err error
	config, err = GetConfig()
	if err != nil {
		log.Fatal(err)
	}
}

// Review do multiple reviews
func Review(pr *PullRequest) error {
	diffs, err := getDiffsBetweenBranches(getBranchNameFromRefName(pr.Resource.TargetRefName), getBranchNameFromRefName(pr.Resource.SourceRefName))
	if err != nil {
		return err
	}

	ir := imageReview{diffs, pr}
	imagePass, err := ir.review()
	if err != nil {
		return err
	}

	cgr := changeGroupReview{diffs, pr}
	changeGroupPass, err := cgr.review()
	if err != nil {
		return err
	}

	ser := storageEntitiesReview{diffs, pr}
	storageEntitiesPass, err := ser.review()
	if err != nil {
		return err
	}

	err = vote(pr, imagePass && changeGroupPass && storageEntitiesPass)
	if err != nil {
		return err
	}

	return nil
}

func vote(pr *PullRequest, pass bool) error {
	if pass {
		for _, reviewer := range pr.Resource.Reviewers {
			if strings.EqualFold(reviewer.ID, config.UserID) {
				if reviewer.Vote < 0 {
					// reset
					err := votePullRequest(pr.Resource.PullRequestID, 0)
					if err != nil {
						return err
					}
				}
				break
			}
		}
	} else {
		// wait
		err := votePullRequest(pr.Resource.PullRequestID, -5)
		if err != nil {
			return err
		}
	}

	return nil
}
