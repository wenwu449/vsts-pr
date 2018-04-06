package vsts

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// PullRequest is a pull request from VSTS
type PullRequest struct {
	ID          string `json:"id"`
	EventType   string `json:"eventType"`
	PublisherID string `json:"publisherId"`
	Resource    struct {
		Repository struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			URL     string `json:"url"`
			Project struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				URL        string `json:"url"`
				State      string `json:"state"`
				Revision   int    `json:"revision"`
				Visibility string `json:"visibility"`
			} `json:"project"`
			RemoteURL string `json:"remoteUrl"`
			SSHURL    string `json:"sshUrl"`
		} `json:"repository"`
		PullRequestID int    `json:"pullRequestId"`
		CodeReviewID  int    `json:"codeReviewId"`
		Status        string `json:"status"`
		CreatedBy     struct {
			DisplayName string `json:"displayName"`
			URL         string `json:"url"`
			ID          string `json:"id"`
			UniqueName  string `json:"uniqueName"`
			ImageURL    string `json:"imageUrl"`
			Descriptor  string `json:"descriptor"`
		} `json:"createdBy"`
		CreationDate          time.Time `json:"creationDate"`
		Title                 string    `json:"title"`
		Description           string    `json:"description"`
		SourceRefName         string    `json:"sourceRefName"`
		TargetRefName         string    `json:"targetRefName"`
		MergeStatus           string    `json:"mergeStatus"`
		MergeID               string    `json:"mergeId"`
		LastMergeSourceCommit struct {
			CommitID string `json:"commitId"`
			URL      string `json:"url"`
		} `json:"lastMergeSourceCommit"`
		LastMergeTargetCommit struct {
			CommitID string `json:"commitId"`
			URL      string `json:"url"`
		} `json:"lastMergeTargetCommit"`
		LastMergeCommit struct {
			CommitID string `json:"commitId"`
			Author   struct {
				Name  string    `json:"name"`
				Email string    `json:"email"`
				Date  time.Time `json:"date"`
			} `json:"author"`
			Committer struct {
				Name  string    `json:"name"`
				Email string    `json:"email"`
				Date  time.Time `json:"date"`
			} `json:"committer"`
			Comment string `json:"comment"`
			URL     string `json:"url"`
		} `json:"lastMergeCommit"`
		Reviewers []struct {
			ReviewerURL string `json:"reviewerUrl"`
			Vote        int    `json:"vote"`
			DisplayName string `json:"displayName"`
			URL         string `json:"url"`
			ID          string `json:"id"`
			UniqueName  string `json:"uniqueName"`
			ImageURL    string `json:"imageUrl"`
			IsContainer bool   `json:"isContainer,omitempty"`
			VotedFor    []struct {
				ReviewerURL string `json:"reviewerUrl"`
				Vote        int    `json:"vote"`
				DisplayName string `json:"displayName"`
				URL         string `json:"url"`
				ID          string `json:"id"`
				UniqueName  string `json:"uniqueName"`
				ImageURL    string `json:"imageUrl"`
				IsContainer bool   `json:"isContainer"`
			} `json:"votedFor,omitempty"`
		} `json:"reviewers"`
		URL   string `json:"url"`
		Links struct {
			Web struct {
				Href string `json:"href"`
			} `json:"web"`
			Statuses struct {
				Href string `json:"href"`
			} `json:"statuses"`
		} `json:"_links"`
		SupportsIterations bool   `json:"supportsIterations"`
		ArtifactID         string `json:"artifactId"`
	} `json:"resource"`
	ResourceVersion    string `json:"resourceVersion"`
	ResourceContainers struct {
		Collection struct {
			ID      string `json:"id"`
			BaseURL string `json:"baseUrl"`
		} `json:"collection"`
		Account struct {
			ID      string `json:"id"`
			BaseURL string `json:"baseUrl"`
		} `json:"account"`
		Project struct {
			ID      string `json:"id"`
			BaseURL string `json:"baseUrl"`
		} `json:"project"`
	} `json:"resourceContainers"`
	CreatedDate time.Time `json:"createdDate"`
}

// ParsePullRequest parse pull request from encoded string
func ParsePullRequest() (*PullRequest, error) {
	encodedPRContentString := os.Getenv("PR_CONTENT")
	if len(encodedPRContentString) == 0 {
		return nil, fmt.Errorf("env PR_CONTENT not found")
	}

	prContentBytes, err := base64.StdEncoding.DecodeString(encodedPRContentString)
	if err != nil {
		return nil, err
	}
	prContentString := string(prContentBytes)
	prContentRaw := prContentString[strings.Index(prContentString, "{"):(strings.LastIndex(prContentString, "}") + 1)]
	prContent := PullRequest{}
	if err := json.Unmarshal([]byte(prContentRaw), &prContent); err != nil {
		return nil, err
	}

	if strings.EqualFold(prContent.ID, "") {
		return nil, fmt.Errorf("PR ID is empty: %v", prContentString)
	}

	return &prContent, nil
}
