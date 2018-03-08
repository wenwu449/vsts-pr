package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type imageConfig struct {
	Os         string `json:"os"`
	ConfigPath string `json:"configPath"`
	Header     string `json:"header"`
}

type secrets struct {
	Username     string        `json:"username"`
	Password     string        `json:"password"`
	Instance     string        `json:"instance"`
	Collection   string        `json:"collection"`
	Project      string        `json:"project"`
	Repo         string        `json:"repo"`
	MasterBranch string        `json:"masterBranch"`
	UserID       string        `json:"userId"`
	ImageConfigs []imageConfig `json:"imageConfigs"`
	Endpoints    []string      `json:"endpoints"`
}

type prcontent struct {
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

type imageList struct {
	Version string   `json:"version"`
	Common  []string `json:"common"`
}

type author struct {
	ID string `json:"id"`
}

type comment struct {
	ID              int    `json:"id"`
	ParentCommentID int    `json:"parentCommentId"`
	Author          author `json:"author"`
	Content         string `json:"content"`
	CommentType     string `json:"commentType"`
	IsDeleted       bool   `json:"isDeleted"`
}

type filePosition struct {
	Line   int `json:"line"`
	Offset int `json:"offset"`
}

type threadContext struct {
	FilePath       string       `json:"filePath"`
	LeftFileStart  filePosition `json:"leftFileStart"`
	LeftFileEnd    filePosition `json:"leftFileEnd"`
	RightFileStart filePosition `json:"rightFileStart"`
	RightFileEnd   filePosition `json:"rightFileEnd"`
}

type commentThread struct {
	ID            int           `json:"id"`
	Comments      []comment     `json:"comments"`
	Status        string        `json:"status"`
	ThreadContext threadContext `json:"threadContext"`
	IsDeleted     bool          `json:"isDeleted"`
}

type commentThreads struct {
	Value []commentThread `json:"value"`
	Count int             `json:"count"`
}

type postComment struct {
	ParentCommentID int    `json:"parentCommentId"`
	Content         string `json:"content"`
	CommentType     int    `json:"commentType"`
}

type supportsMarkDown struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type threadProperty struct {
	MicrosoftTeamFoundationDiscussionSupportsMarkdown supportsMarkDown `json:"Microsoft.TeamFoundation.Discussion.SupportsMarkdown"`
}

type postThread struct {
	Comments      []postComment  `json:"comments"`
	Properties    threadProperty `json:"properties"`
	Status        int            `json:"status"`
	ThreadContext threadContext  `json:"threadContext"`
}

type patchThread struct {
	Status int `json:"status"`
}

type putVote struct {
	Vote int `json:"vote"`
}

type diffs struct {
	AllChangesIncluded bool `json:"allChangesIncluded"`
	ChangeCounts       struct {
		Edit int `json:"Edit"`
	} `json:"changeCounts"`
	Changes []struct {
		Item struct {
			ObjectID         string `json:"objectId"`
			OriginalObjectID string `json:"originalObjectId"`
			GitObjectType    string `json:"gitObjectType"`
			CommitID         string `json:"commitId"`
			Path             string `json:"path"`
			IsFolder         bool   `json:"isFolder"`
			URL              string `json:"url"`
		} `json:"item"`
		ChangeType string `json:"changeType"`
	} `json:"changes"`
	CommonCommit string `json:"commonCommit"`
	BaseCommit   string `json:"baseCommit"`
	TargetCommit string `json:"targetCommit"`
	AheadCount   int    `json:"aheadCount"`
	BehindCount  int    `json:"behindCount"`
}

var secret = secrets{}
var botCommentPrefix = "**[BOT_Image]**"
var botCommentSuffix = "*This comment was added by a bot, please kindly let me know if the comment needs improvement.*"

func getBranchNameFromRefName(refName string) string {
	return (strings.Split(refName, "/"))[2]
}

func getDiffsBetweenBranches(client *http.Client, baseBranch string, targetBranch string) diffs {
	getDiffsURLTemplate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/diffs/commits?api-version={version}&targetVersionType=branch&targetVersion={targetBranch}&baseVersionType=branch&baseVersion={baseBranch}"
	r := strings.NewReplacer(
		"{instance}", secret.Instance,
		"{project}", secret.Project,
		"{repository}", secret.Repo,
		"{version}", "1.0",
		"{baseBranch}", baseBranch,
		"{targetBranch}", targetBranch)

	urlString := r.Replace(getDiffsURLTemplate)

	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(secret.Username, secret.Password)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	diffs := diffs{}

	json.NewDecoder(resp.Body).Decode(&diffs)

	return diffs
}

func getBranchItemContent(client *http.Client, branch string, itemPath string, v interface{}) error {
	getItemURLTemplate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/items?api-version={version}&versionType={versionType}&version={versionValue}&scopePath={itemPath}&lastProcessedChange=true"
	r := strings.NewReplacer(
		"{instance}", secret.Instance,
		"{project}", secret.Project,
		"{repository}", secret.Repo,
		"{versionType}", "branch",
		"{versionValue}", branch,
		"{itemPath}", itemPath,
		"{version}", "1.0")

	urlString := r.Replace(getItemURLTemplate)

	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		log.Fatal(err)
		return err
	}

	req.SetBasicAuth(secret.Username, secret.Password)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer resp.Body.Close()

	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = json.Unmarshal(bodyText, v)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func getHeaderFromHealthCheck(done chan<- http.Header, endpoint string) {
	urlString := fmt.Sprintf("%s/%s", endpoint, "healthcheck")
	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		log.Fatal(err)
		done <- nil
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		done <- nil
	}
	done <- resp.Header
}

func getCommentThreads(client *http.Client, pullRequestID int) commentThreads {
	getThreadsURLTemplate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/pullRequests/{pullRequest}/threads?api-version={version}"
	r := strings.NewReplacer(
		"{instance}", secret.Instance,
		"{project}", secret.Project,
		"{repository}", secret.Repo,
		"{pullRequest}", strconv.Itoa(pullRequestID),
		"{version}", "3.0-preview")

	urlString := r.Replace(getThreadsURLTemplate)

	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(secret.Username, secret.Password)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	commentThreads := commentThreads{}
	err = json.Unmarshal(bodyText, &commentThreads)
	if err != nil {
		log.Fatal(err)
	}
	return commentThreads
}

func createCommentThread(client *http.Client, pullRequestID int, filePath string, missingImages []string) {
	fmt.Printf("Creating comment thread to PR %v...\n", pullRequestID)
	postThreadURLTempate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/pullRequests/{pullRequest}/threads?api-version={version}"
	r := strings.NewReplacer(
		"{instance}", secret.Instance,
		"{project}", secret.Project,
		"{repository}", secret.Repo,
		"{pullRequest}", strconv.Itoa(pullRequestID),
		"{version}", "3.0-preview")

	thread := postThread{
		Comments: []postComment{
			{
				ParentCommentID: 0,
				Content:         fmt.Sprintf("%s %s%+v\n%s", botCommentPrefix, "Missing images:", missingImages, botCommentSuffix),
				CommentType:     1,
			},
		},
		Properties: threadProperty{
			MicrosoftTeamFoundationDiscussionSupportsMarkdown: supportsMarkDown{
				Type:  "System.Int32",
				Value: 1,
			},
		},
		Status: 1,
		ThreadContext: threadContext{
			FilePath: filePath,
			LeftFileStart: filePosition{
				Line:   4,
				Offset: 9,
			},
			LeftFileEnd: filePosition{
				Line:   4,
				Offset: 10,
			},
			RightFileStart: filePosition{
				Line:   4,
				Offset: 9,
			},
			RightFileEnd: filePosition{
				Line:   4,
				Offset: 10,
			},
		},
	}

	urlString := r.Replace(postThreadURLTempate)

	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(thread)
	req, err := http.NewRequest("POST", urlString, body)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(secret.Username, secret.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp.Status)
}

func addComment(client *http.Client, pullRequestID int, thread commentThread, missingImages []string) {
	postCommentURLTempate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/pullRequests/{pullRequest}/threads/{threadID}/comments?api-version={version}"
	r := strings.NewReplacer(
		"{instance}", secret.Instance,
		"{project}", secret.Project,
		"{repository}", secret.Repo,
		"{pullRequest}", strconv.Itoa(pullRequestID),
		"{threadID}", strconv.Itoa(thread.ID),
		"{version}", "3.0-preview")

	lastCommentID := 0
	commentContent := ""
	for _, comment := range thread.Comments {
		if !comment.IsDeleted && comment.ID > lastCommentID {
			lastCommentID = comment.ID
			commentContent = comment.Content
		}
	}

	content := fmt.Sprintf("%s %s%+v\n%s", botCommentPrefix, "Missing images:", missingImages, botCommentSuffix)

	if strings.EqualFold(commentContent, content) {
		return
	}

	fmt.Printf("Adding comment to PR %v thread %v...\n", pullRequestID, thread.ID)

	comment := postComment{
		ParentCommentID: lastCommentID,
		Content:         content,
		CommentType:     1,
	}

	urlString := r.Replace(postCommentURLTempate)

	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(comment)
	req, err := http.NewRequest("POST", urlString, body)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(secret.Username, secret.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp.Status)
}

func setCommentThreadActive(client *http.Client, pullRequestID int, thread commentThread) {
	if strings.EqualFold(thread.Status, "active") {
		return
	}

	fmt.Printf("Set PR %v thread %v to active...\n", pullRequestID, thread.ID)

	patchThreadURLTempate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/pullRequests/{pullRequest}/threads/{threadID}?api-version={version}"
	r := strings.NewReplacer(
		"{instance}", secret.Instance,
		"{project}", secret.Project,
		"{repository}", secret.Repo,
		"{pullRequest}", strconv.Itoa(pullRequestID),
		"{threadID}", strconv.Itoa(thread.ID),
		"{version}", "3.0-preview")

	patchThread := patchThread{
		Status: 1,
	}

	urlString := r.Replace(patchThreadURLTempate)

	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(patchThread)
	req, err := http.NewRequest("PATCH", urlString, body)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(secret.Username, secret.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp.Status)
}

func votePullRequest(client *http.Client, pullRequestID int, vote int) {
	fmt.Printf("Vote on PR %v: %v...\n", pullRequestID, vote)

	voteURLTemplate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/pullRequests/{pullRequest}/reviewers/{reviewer}?api-version={version}"
	r := strings.NewReplacer(
		"{instance}", secret.Instance,
		"{project}", secret.Project,
		"{repository}", secret.Repo,
		"{pullRequest}", strconv.Itoa(pullRequestID),
		"{reviewer}", secret.UserID,
		"{version}", "3.0-preview")

	putVote := putVote{
		Vote: vote,
	}

	urlString := r.Replace(voteURLTemplate)

	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(putVote)
	req, err := http.NewRequest("PUT", urlString, body)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(secret.Username, secret.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp.Status)
}

func main() {
	// read secrets
	file, _ := os.Open("secrets.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&secret)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(secret.Username)

	// read PR content
	encodedPRContentString := os.Getenv("PRCONTENT")
	if len(encodedPRContentString) == 0 {
		fmt.Println("env PRCONTENT not found.")
		return
	}

	prContentBytes, err := base64.StdEncoding.DecodeString(encodedPRContentString)
	if err != nil {
		log.Fatal(err)
	}
	prContentString := string(prContentBytes)
	prContentRaw := prContentString[strings.Index(prContentString, "{"):(strings.LastIndex(prContentString, "}") + 1)]
	prContent := prcontent{}
	if err := json.Unmarshal([]byte(prContentRaw), &prContent); err != nil {
		log.Fatal(err)
	}

	if strings.EqualFold(prContent.ID, "") {
		fmt.Printf("unexpected PR content: %v\n", prContentString)
		return
	}
	fmt.Printf("Got PR update: %s\n", prContent.ID)

	if !strings.EqualFold(prContent.Resource.TargetRefName, fmt.Sprintf("%s/%s", "refs/heads", secret.MasterBranch)) {
		fmt.Printf("unexpected target branch: %s\n", prContent.Resource.TargetRefName)
		return
	}

	client := &http.Client{}

	diffs := getDiffsBetweenBranches(client, getBranchNameFromRefName(prContent.Resource.TargetRefName), getBranchNameFromRefName(prContent.Resource.SourceRefName))

	var changedImageConfigs []imageConfig
	for _, imageConfig := range secret.ImageConfigs {
		for _, change := range diffs.Changes {
			if strings.EqualFold(imageConfig.ConfigPath, change.Item.Path) {
				changedImageConfigs = append(changedImageConfigs, imageConfig)
				break
			}
		}
	}

	if len(changedImageConfigs) == 0 {
		fmt.Println("No change in image config")
		return
	}

	imageDistinct := make(map[string]map[string]struct{})
	for _, imageConfig := range secret.ImageConfigs {
		imageDistinct[imageConfig.Os] = make(map[string]struct{})
	}

	done := make(chan http.Header)
	for _, endpoint := range secret.Endpoints {
		go getHeaderFromHealthCheck(done, endpoint)
	}

	for _ = range secret.Endpoints {
		header := <-done
		if header != nil {
			for _, imageConfig := range secret.ImageConfigs {
				imageVersion := header.Get(imageConfig.Header)
				if _, ok := imageDistinct[imageConfig.Os][imageVersion]; !ok {
					imageDistinct[imageConfig.Os][imageVersion] = struct{}{}
				}
			}
		}
	}

	missingImages := make(map[string][]string)
	for _, imageConfig := range changedImageConfigs {
		images := []string{}
		imageList := imageList{}
		err := getBranchItemContent(client, getBranchNameFromRefName(prContent.Resource.SourceRefName), imageConfig.ConfigPath, &imageList)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Checking: %s\n", imageConfig.ConfigPath)
		fmt.Printf("images: %+v\n", imageList.Common)
		requiredImages := imageDistinct[imageConfig.Os]
		for imageVersion := range requiredImages {
			fmt.Printf("Checking required image: %s\n", imageVersion)
			found := false
			for _, image := range imageList.Common {
				version := image[strings.LastIndex(image, ":")+1:]
				if strings.EqualFold(version, imageVersion) {
					found = true
					break
				}
			}
			if !found {
				images = append(images, imageVersion)
				fmt.Printf("%s missing %s\n", imageConfig.ConfigPath, imageVersion)
			}
		}
		if len(images) > 0 {
			missingImages[imageConfig.ConfigPath] = images
		}
	}

	if len(missingImages) == 0 {
		foundVote := false
		for _, reviewer := range prContent.Resource.Reviewers {
			if strings.EqualFold(reviewer.ID, secret.UserID) {
				foundVote = true
				if reviewer.Vote < 0 {
					votePullRequest(client, prContent.Resource.PullRequestID, 0)
				}
				break
			}
		}
		if !foundVote {
			votePullRequest(client, prContent.Resource.PullRequestID, 0)
		}
	}

	fmt.Printf("image missing: %+v\n", missingImages)

	commentThreads := getCommentThreads(client, prContent.Resource.PullRequestID)
	for filePath, missingImages := range missingImages {
		commentThread := commentThread{}
		for _, thread := range commentThreads.Value {
			if !thread.IsDeleted && strings.EqualFold(thread.ThreadContext.FilePath, filePath) {
				for _, comment := range thread.Comments {
					if comment.ID == 1 && comment.Author.ID == secret.UserID && strings.HasPrefix(comment.Content, botCommentPrefix) {
						commentThread = thread
						break
					}
				}
			}
		}
		if commentThread.Status == "" {
			// create thread
			createCommentThread(client, prContent.Resource.PullRequestID, filePath, missingImages)
		} else {
			// add comment
			addComment(client, prContent.Resource.PullRequestID, commentThread, missingImages)
			// set active
			setCommentThreadActive(client, prContent.Resource.PullRequestID, commentThread)
		}
	}
	// wait
	votePullRequest(client, prContent.Resource.PullRequestID, -5)
}
