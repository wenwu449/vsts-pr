package vsts

import (
	"log"
	"strconv"
	"strings"
)

func getThreadsURL(pullRequestID int) string {
	threadsURLTemplate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/pullRequests/{pullRequest}/threads?api-version={version}"

	r := strings.NewReplacer(
		"{instance}", config.Instance,
		"{project}", config.Project,
		"{repository}", config.Repo,
		"{pullRequest}", strconv.Itoa(pullRequestID),
		"{version}", "3.0-preview")

	return r.Replace(threadsURLTemplate)
}

func getThreadURL(pullRequestID int, threadID int) string {
	threadURLTempate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/pullRequests/{pullRequest}/threads/{threadID}?api-version={version}"
	r := strings.NewReplacer(
		"{instance}", config.Instance,
		"{project}", config.Project,
		"{repository}", config.Repo,
		"{pullRequest}", strconv.Itoa(pullRequestID),
		"{threadID}", strconv.Itoa(threadID),
		"{version}", "3.0-preview")

	return r.Replace(threadURLTempate)
}

func getCommentURL(pullRequestID int, threadID int) string {
	commentURLTempate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/pullRequests/{pullRequest}/threads/{threadID}/comments?api-version={version}"
	r := strings.NewReplacer(
		"{instance}", config.Instance,
		"{project}", config.Project,
		"{repository}", config.Repo,
		"{pullRequest}", strconv.Itoa(pullRequestID),
		"{threadID}", strconv.Itoa(threadID),
		"{version}", "3.0-preview")

	return r.Replace(commentURLTempate)
}

func getCommentThreads(pullRequestID int) (*commentThreads, error) {
	commentThreads := new(commentThreads)

	url := getThreadsURL(pullRequestID)
	err := getFromVsts(url, commentThreads)

	if err != nil {
		return nil, err
	}

	return commentThreads, nil
}

func createCommentThread(pullRequestID int, filePath string, status int, content string) error {
	log.Printf("Creating comment thread to PR %v...\n", pullRequestID)

	thread := postThread{
		Comments: []postComment{
			{
				ParentCommentID: 0,
				Content:         content,
				CommentType:     1,
			},
		},
		Properties: threadProperty{
			MicrosoftTeamFoundationDiscussionSupportsMarkdown: supportsMarkDown{
				Type:  "System.Int32",
				Value: 1,
			},
		},
		Status: status,
		ThreadContext: threadContext{
			FilePath: filePath,
			RightFileStart: filePosition{
				Line:   1,
				Offset: 1,
			},
			RightFileEnd: filePosition{
				Line:   1,
				Offset: 3,
			},
		},
	}

	if filePath == "" {
		thread.ThreadContext = threadContext{}
	}

	url := getThreadsURL(pullRequestID)

	err := postToVsts(url, thread)
	if err != nil {
		return err
	}

	return nil
}

func addComment(pullRequestID int, thread commentThread, essentialMessage string, content string) error {
	lastCommentID := 0
	commentContent := ""
	for _, comment := range thread.Comments {
		if !comment.IsDeleted && comment.ID > lastCommentID {
			lastCommentID = comment.ID
			commentContent = comment.Content
		}
	}

	if strings.Contains(commentContent, essentialMessage) {
		log.Printf("Already commented to PR %v thread %v...\n", pullRequestID, thread.ID)
		return nil
	}

	log.Printf("Adding comment to PR %v thread %v...\n", pullRequestID, thread.ID)

	comment := postComment{
		ParentCommentID: lastCommentID,
		Content:         content,
		CommentType:     1,
	}

	url := getCommentURL(pullRequestID, thread.ID)

	err := postToVsts(url, comment)
	if err != nil {
		return err
	}

	return nil
}

func setCommentThreadStatus(pullRequestID int, thread commentThread, status int) error {
	statusString := "active"
	if status == 2 {
		statusString = "fixed"
	}
	if strings.EqualFold(thread.Status, statusString) {
		log.Printf("PR %v thread %v status is already %v\n", pullRequestID, thread.ID, status)
		return nil
	}

	log.Printf("Set PR %v thread %v to %v...\n", pullRequestID, thread.ID, status)

	patchThread := patchThread{
		Status: status,
	}

	url := getThreadURL(pullRequestID, thread.ID)

	err := patchToVsts(url, patchThread)
	if err != nil {
		return err
	}

	return nil
}
