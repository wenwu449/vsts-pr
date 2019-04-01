package vsts

import (
	"log"
	"strconv"
	"strings"
)

type Commenter struct {
	config *Config
	client *Client
}

func NewCommenter(config *Config, client *Client) *Commenter {
	return &Commenter{
		config: config,
		client: client,
	}
}

func (c *Commenter) getThreadsURL(pullRequestID int) string {
	threadsURLTemplate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/pullRequests/{pullRequest}/threads?api-version={version}"

	r := strings.NewReplacer(
		"{instance}", c.config.Instance,
		"{project}", c.config.Project,
		"{repository}", c.config.Repo,
		"{pullRequest}", strconv.Itoa(pullRequestID),
		"{version}", "3.0-preview")

	return r.Replace(threadsURLTemplate)
}

func (c *Commenter) getThreadURL(pullRequestID int, threadID int) string {
	threadURLTempate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/pullRequests/{pullRequest}/threads/{threadID}?api-version={version}"
	r := strings.NewReplacer(
		"{instance}", c.config.Instance,
		"{project}", c.config.Project,
		"{repository}", c.config.Repo,
		"{pullRequest}", strconv.Itoa(pullRequestID),
		"{threadID}", strconv.Itoa(threadID),
		"{version}", "3.0-preview")

	return r.Replace(threadURLTempate)
}

func (c *Commenter) getCommentURL(pullRequestID int, threadID int) string {
	commentURLTempate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/pullRequests/{pullRequest}/threads/{threadID}/comments?api-version={version}"
	r := strings.NewReplacer(
		"{instance}", c.config.Instance,
		"{project}", c.config.Project,
		"{repository}", c.config.Repo,
		"{pullRequest}", strconv.Itoa(pullRequestID),
		"{threadID}", strconv.Itoa(threadID),
		"{version}", "3.0-preview")

	return r.Replace(commentURLTempate)
}

func (c *Commenter) getCommentThreads(pullRequestID int) (*commentThreads, error) {
	commentThreads := new(commentThreads)

	url := c.getThreadsURL(pullRequestID)
	err := c.client.getFromVsts(url, commentThreads)

	if err != nil {
		return nil, err
	}

	return commentThreads, nil
}

func (c *Commenter) createCommentThread(pullRequestID int, filePath string, status int, content string) error {
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

	url := c.getThreadsURL(pullRequestID)

	err := c.client.postToVsts(url, thread)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commenter) addComment(pullRequestID int, thread commentThread, essentialMessage string, content string) error {
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

	url := c.getCommentURL(pullRequestID, thread.ID)

	err := c.client.postToVsts(url, comment)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commenter) setCommentThreadStatus(pullRequestID int, thread commentThread, status int) error {
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

	url := c.getThreadURL(pullRequestID, thread.ID)

	err := c.client.patchToVsts(url, patchThread)
	if err != nil {
		return err
	}

	return nil
}
