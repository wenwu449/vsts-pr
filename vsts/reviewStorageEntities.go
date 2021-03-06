package vsts

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

type storageEntitiesReview struct {
	diffs       *diffs
	pullRequest *PullRequest
}

func (r *storageEntitiesReview) getBotCommentPrefix() string {
	return "[BOT_Entities]\n"
}

func (r *storageEntitiesReview) getBotCommentSuffix() string {
	return "\n*This comment was added by bot, please let me know if you have any suggestion!*"
}

func (r *storageEntitiesReview) getCommentContent(changedStorageEntityPathes []string) string {
	return fmt.Sprintf(
		"%s:warning:\nThe following storage entities were changed:\n**%+v**\nMake sure you are not removing any properties that will break back compatibility.%s",
		r.getBotCommentPrefix(),
		changedStorageEntityPathes,
		r.getBotCommentSuffix())
}

func (r *storageEntitiesReview) review() (bool, error) {
	log.Println("storage entities check started.")

	var changedStorageEntityPathes []string
	for _, change := range r.diffs.Changes {
		for _, storageEntityPrefix := range config.StorageEntitiesPrefix {
			// Ignore folders.
			// Usually add new entities won't break back compatibility, thus ignore.
			if change.ChangeType != "add" && !change.Item.IsFolder && strings.HasPrefix(change.Item.Path, storageEntityPrefix) {
				changedStorageEntityPathes = append(changedStorageEntityPathes, change.Item.Path)
			}
		}
	}

	sort.Strings(changedStorageEntityPathes)

	if len(changedStorageEntityPathes) == 0 {
		log.Printf("storage entities check passed.\n")
		return true, nil
	}

	log.Printf("storage entities check contains warning for files: %+v\n", changedStorageEntityPathes)

	commentThreads, err := getCommentThreads(r.pullRequest.Resource.PullRequestID)
	if err != nil {
		return false, err
	}

	commentThread := commentThread{}
	for _, thread := range commentThreads.Value {
		if !thread.IsDeleted && thread.ThreadContext.FilePath == "" {
			for _, comment := range thread.Comments {
				if comment.ID == 1 && comment.Author.ID == config.UserID && strings.HasPrefix(comment.Content, r.getBotCommentPrefix()) {
					commentThread = thread
					break
				}
			}
		}
	}

	if commentThread.Status == "" {
		commentContent := r.getCommentContent(changedStorageEntityPathes)
		err := createCommentThread(r.pullRequest.Resource.PullRequestID, "", 1, commentContent)
		if err != nil {
			return false, err
		}

		// Only return false when creating the comment for the first time.
		return false, nil
	}

	// As long as the comment exists, just pass.
	log.Printf("storage entities check completed.\n")
	return true, nil
}
