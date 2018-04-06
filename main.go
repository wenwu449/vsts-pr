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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wenwu449/vsts-pr/vsts"
)




// type comment struct {
// 	ID              int    `json:"id"`
// 	ParentCommentID int    `json:"parentCommentId"`
// 	Author          author `json:"author"`
// 	Content         string `json:"content"`
// 	CommentType     string `json:"commentType"`
// 	IsDeleted       bool   `json:"isDeleted"`
// }

// type filePosition struct {
// 	Line   int `json:"line"`
// 	Offset int `json:"offset"`
// }

// type threadContext struct {
// 	FilePath       string       `json:"filePath"`
// 	LeftFileStart  filePosition `json:"leftFileStart"`
// 	LeftFileEnd    filePosition `json:"leftFileEnd"`
// 	RightFileStart filePosition `json:"rightFileStart"`
// 	RightFileEnd   filePosition `json:"rightFileEnd"`
// }

// type commentThread struct {
// 	ID            int           `json:"id"`
// 	Comments      []comment     `json:"comments"`
// 	Status        string        `json:"status"`
// 	ThreadContext threadContext `json:"threadContext"`
// 	IsDeleted     bool          `json:"isDeleted"`
// }

// type commentThreads struct {
// 	Value []commentThread `json:"value"`
// 	Count int             `json:"count"`
// }

// type postComment struct {
// 	ParentCommentID int    `json:"parentCommentId"`
// 	Content         string `json:"content"`
// 	CommentType     int    `json:"commentType"`
// }

// type supportsMarkDown struct {
// 	Type  string `json:"type"`
// 	Value int    `json:"value"`
// }

// type threadProperty struct {
// 	MicrosoftTeamFoundationDiscussionSupportsMarkdown supportsMarkDown `json:"Microsoft.TeamFoundation.Discussion.SupportsMarkdown"`
// }

// type postThread struct {
// 	Comments      []postComment  `json:"comments"`
// 	Properties    threadProperty `json:"properties"`
// 	Status        int            `json:"status"`
// 	ThreadContext threadContext  `json:"threadContext"`
// }

// type patchThread struct {
// 	Status int `json:"status"`
// }

// type putVote struct {
// 	Vote int `json:"vote"`
// }

// type diffs struct {
// 	AllChangesIncluded bool `json:"allChangesIncluded"`
// 	ChangeCounts       struct {
// 		Edit int `json:"Edit"`
// 	} `json:"changeCounts"`
// 	Changes []struct {
// 		Item struct {
// 			ObjectID         string `json:"objectId"`
// 			OriginalObjectID string `json:"originalObjectId"`
// 			GitObjectType    string `json:"gitObjectType"`
// 			CommitID         string `json:"commitId"`
// 			Path             string `json:"path"`
// 			IsFolder         bool   `json:"isFolder"`
// 			URL              string `json:"url"`
// 		} `json:"item"`
// 		ChangeType string `json:"changeType"`
// 	} `json:"changes"`
// 	CommonCommit string `json:"commonCommit"`
// 	BaseCommit   string `json:"baseCommit"`
// 	TargetCommit string `json:"targetCommit"`
// 	AheadCount   int    `json:"aheadCount"`
// 	BehindCount  int    `json:"behindCount"`
// }

var secret = secrets{}
var botCommentPrefix = "[BOT_Image]\n"
var botCommentSuffix = "\n*This comment was added by bot, please let me know if you have any suggestion!*"

func getBranchNameFromRefName(refName string) string {
	return (strings.SplitAfterN(refName, "/", 3))[2]
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

func getFailedSign() string {
	signs := []string{":exclamation:", ":warning:", ":scream:", ":broken_heart:", ":no_entry:", ":bangbang:"}
	return signs[time.Now().Minute()%len(signs)]
}

func getPassedSign() string {
	signs := []string{":heavy_check_mark:", ":thumbsup:", ":white_check_mark:", ":ballot_box_with_check:", ":clap:", ":smiley:"}
	return signs[time.Now().Minute()%len(signs)]
}

func getPassedWord() string {
	words := []string{"Verified!", "Awesome!", "Clear!", "Great Job!"}
	return words[time.Now().Minute()%len(words)]
}

func getCommentContent(missingImages []string) (string, string) {
	essentialMessage := "All images listed on [acchealth](http://armhealth.azurewebsites.net/#acc) are included in this list."
	if len(missingImages) == 0 {
		return essentialMessage, fmt.Sprintf(
			"%s%s %s %s\n%s",
			botCommentPrefix,
			getPassedSign(),
			getPassedWord(),
			essentialMessage,
			botCommentSuffix)
	}
	sort.Strings(missingImages)
	essentialMessage = fmt.Sprintf("Following images should be included:**%+v**", missingImages)
	return essentialMessage, fmt.Sprintf(
		"%s%s %s\nYou can get image list from [acchealth](http://armhealth.azurewebsites.net/#acc).\n%s",
		botCommentPrefix,
		getFailedSign(),
		essentialMessage,
		botCommentSuffix)
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

	status := 1
	if len(missingImages) == 0 {
		status = 2
	}

	_, content := getCommentContent(missingImages)
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

	essentialMessage, content := getCommentContent(missingImages)

	if strings.Contains(commentContent, essentialMessage) {
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

func setCommentThreadStatus(client *http.Client, pullRequestID int, thread commentThread, status int) {
	statusString := "active"
	if status == 2 {
		statusString = "fixed"
	}
	if strings.EqualFold(thread.Status, statusString) {
		fmt.Printf("thread status is already %v\n", status)
		return
	}

	fmt.Printf("Set PR %v thread %v to %v...\n", pullRequestID, thread.ID, status)

	patchThreadURLTempate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/pullRequests/{pullRequest}/threads/{threadID}?api-version={version}"
	r := strings.NewReplacer(
		"{instance}", secret.Instance,
		"{project}", secret.Project,
		"{repository}", secret.Repo,
		"{pullRequest}", strconv.Itoa(pullRequestID),
		"{threadID}", strconv.Itoa(thread.ID),
		"{version}", "3.0-preview")

	patchThread := patchThread{
		Status: status,
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
	log.SetFlags(log.LstdFlags | log.Llongfile)

	config, err := vsts.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	pr, err := vsts.ParsePullRequest()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Got PR update: %s\n", pr.ID)

	if !strings.EqualFold(prContent.Resource.TargetRefName, fmt.Sprintf("%s/%s", "refs/heads", secret.MasterBranch)) {
		fmt.Printf("unexpected target branch: %s\n", prContent.Resource.TargetRefName)
		return
	}

	client := &http.Client{}

	diffs := getDiffsBetweenBranches(client, getBranchNameFromRefName(prContent.Resource.TargetRefName), getBranchNameFromRefName(prContent.Resource.SourceRefName))

	// image check
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

	missingImagesMap := make(map[string][]string)
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
			missingImagesMap[imageConfig.ConfigPath] = images
		}
	}

	if len(missingImagesMap) == 0 {
		fmt.Printf("image check pass.\n")
	} else {
		fmt.Printf("image missing: %+v\n", missingImagesMap)
	}

	commentThreads := getCommentThreads(client, prContent.Resource.PullRequestID)
	for _, imageConfig := range changedImageConfigs {
		missingImages, ok := missingImagesMap[imageConfig.ConfigPath]
		if !ok {
			missingImages = []string{}
		}

		commentThread := commentThread{}
		for _, thread := range commentThreads.Value {
			if !thread.IsDeleted && strings.EqualFold(thread.ThreadContext.FilePath, imageConfig.ConfigPath) {
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
			createCommentThread(client, prContent.Resource.PullRequestID, imageConfig.ConfigPath, missingImages)
		} else {
			// add comment
			addComment(client, prContent.Resource.PullRequestID, commentThread, missingImages)
			if len(missingImages) == 0 {
				// set fixted
				setCommentThreadStatus(client, prContent.Resource.PullRequestID, commentThread, 2)
			} else {
				// set active
				setCommentThreadStatus(client, prContent.Resource.PullRequestID, commentThread, 1)
			}
		}
	}

	// change pair check
	changedItemMap := make(map[string]bool)
	for _, change := range diffs.Changes {
		changedItemMap[change.Item.Path] = true
	}

	missingGroupMap = make(map[string]changeGroup)
	for _, group := range secret.ChangeGroups {
		list := make([]string)
		for _, item := range group {
			if _, ok := changedItemMap[item]; ok {
				list = append(list, item)
			}
		}

		if len(list) > 0 || len(list) < len(group) {
			for _, item := range list {
				missingGroupMap[item] = group
			}
		}
	}

	for filePath, missingGroup := missingGroupMap {
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
			createCommentThread(client, prContent.Resource.PullRequestID, imageConfig.ConfigPath, missingImages)
		} else {
			// add comment
			addComment(client, prContent.Resource.PullRequestID, commentThread, missingImages)
			if len(missingImages) == 0 {
				// set fixted
				setCommentThreadStatus(client, prContent.Resource.PullRequestID, commentThread, 2)
			} else {
				// set active
				setCommentThreadStatus(client, prContent.Resource.PullRequestID, commentThread, 1)
			}
		}
	}

	// vote
	if len(missingImagesMap) == 0 {
		for _, reviewer := range prContent.Resource.Reviewers {
			if strings.EqualFold(reviewer.ID, secret.UserID) {
				if reviewer.Vote < 0 {
					// reset
					votePullRequest(client, prContent.Resource.PullRequestID, 0)
				}
				break
			}
		}
	} else {
		// wait
		votePullRequest(client, prContent.Resource.PullRequestID, -5)
	}

}
