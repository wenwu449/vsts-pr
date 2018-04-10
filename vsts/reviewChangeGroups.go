package vsts

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

type changeGroupReview struct {
	diffs       *diffs
	pullRequest *PullRequest
}

func (r *changeGroupReview) getBotCommentPrefix() string {
	return "[BOT_Group]\n"
}

func (r *changeGroupReview) getBotCommentSuffix() string {
	return "\n*This comment was added by bot, please let me know if you have any suggestion!*"
}

func (r *changeGroupReview) getFailedSign() string {
	signs := []string{":exclamation:", ":warning:", ":scream:", ":broken_heart:", ":no_entry:", ":bangbang:"}
	return signs[time.Now().Minute()%len(signs)]
}

func (r *changeGroupReview) getCommentContent(group changeGroup) (string, string) {
	sort.Strings(group)
	essentialMessage := fmt.Sprintf("These files are usually updated together: **%+v**, please double check.", group)
	return essentialMessage, fmt.Sprintf(
		"%s%s %s\n%s",
		r.getBotCommentPrefix(),
		r.getFailedSign(),
		essentialMessage,
		r.getBotCommentSuffix())
}

func (r *changeGroupReview) review() (bool, error) {
	log.Println("change group check started.")

	changedItemMap := make(map[string]bool)
	for _, change := range r.diffs.Changes {
		changedItemMap[change.Item.Path] = true
	}

	missingGroupMap := make(map[string]changeGroup)
	for _, group := range config.ChangeGroups {
		list := []string{}
		for _, item := range group {
			if _, ok := changedItemMap[item]; ok {
				list = append(list, item)
			}
		}

		if len(list) > 0 && len(list) < len(group) {
			for _, item := range list {
				missingGroupMap[item] = group
			}
		}
	}

	if len(missingGroupMap) == 0 {
		log.Printf("change group check passed.\n")
		return true, nil
	}

	log.Printf("change group failed: %+v\n", missingGroupMap)

	commentThreads, err := getCommentThreads(r.pullRequest.Resource.PullRequestID)
	if err != nil {
		return false, err
	}

	for filePath, missingGroup := range missingGroupMap {
		commentThread := commentThread{}
		for _, thread := range commentThreads.Value {
			if !thread.IsDeleted && strings.EqualFold(thread.ThreadContext.FilePath, filePath) {
				for _, comment := range thread.Comments {
					if comment.ID == 1 && comment.Author.ID == config.UserID && strings.HasPrefix(comment.Content, r.getBotCommentPrefix()) {
						commentThread = thread
						break
					}
				}
			}
		}

		_, commentContent := r.getCommentContent(missingGroup)

		// only add comment once per file.
		if commentThread.Status == "" {
			// create thread
			err := createCommentThread(r.pullRequest.Resource.PullRequestID, filePath, 1, commentContent)
			if err != nil {
				return false, err
			}
		} else {
			log.Printf("Already commented on file %s\n", filePath)
		}
	}

	log.Println("change group completed.")

	return true, nil
}
