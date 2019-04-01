package vsts

import (
	"log"
	"strings"
)

type Reviewer struct {
	config *Config
	client *Client
}

func NewReviewer(config *Config, client *Client) *Reviewer {
	return &Reviewer{
		config: config,
		client: client,
	}
}

// Review do multiple reviews
func (r *Reviewer) Review(pr *PullRequest) error {
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
		log.Printf("All check passed for PR: %v\n", pr.Resource.PullRequestID)

		c, err := containsHumanComments(pr)
		if err != nil {
			return nil
		}
		if c {
			log.Printf("PR contains human comments, skip voting.\n")
			return nil
		}
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

func containsHumanComments(pr *PullRequest) (bool, error) {
	commentThreads, err := getCommentThreads(pr.Resource.PullRequestID)
	if err != nil {
		return false, err
	}

	for _, thread := range commentThreads.Value {
		if !thread.IsDeleted {
			for _, comment := range thread.Comments {
				if !comment.IsDeleted &&
					!strings.EqualFold(comment.CommentType, "system") &&
					strings.EqualFold(comment.Author.ID, config.UserID) &&
					!strings.HasPrefix(comment.Content, "[BOT_") {
					log.Printf("Found human comment: %+v\n", comment)
					return true, nil
				}
			}
		}
	}

	return false, nil
}
