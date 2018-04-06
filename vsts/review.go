package vsts

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	botCommentPrefix = "[BOT]\n"
)

var config *Config

func init() {
	var err error
	config, err = GetConfig()
	if err != nil {
		log.Fatal(err)
	}
}

// ReviewGoTest review if all changed Go files have test updated.
func ReviewGoTest(pr *PullRequest) error {
	goSuffix := ".go"
	goTestSuffix := "_test.go"
	commentMsg := fmt.Sprintf("%s\nPlease update test.", botCommentPrefix)

	diffs, err := getDiffsBetweenBranches(getBranchNameFromRefName(pr.Resource.TargetRefName), getBranchNameFromRefName(pr.Resource.SourceRefName))
	if err != nil {
		return err
	}

	var changedGoFiles []string
	var changedGoTestFiles []string

	for _, change := range diffs.Changes {
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

	commentThreads, err := getCommentThreads(pr.Resource.PullRequestID)
	log.Printf("threads: %v", commentThreads.Count)

	if err != nil {
		return err
	}

	for _, goFile := range missingTestGoFiles {
		commentThread := commentThread{}
		for _, thread := range commentThreads.Value {
			if !thread.IsDeleted && strings.EqualFold(thread.ThreadContext.FilePath, goFile) {
				for _, comment := range thread.Comments {
					if comment.ID == 1 && comment.Author.ID == config.UserID && strings.HasPrefix(comment.Content, botCommentPrefix) {
						commentThread = thread
						break
					}
				}
			}
		}

		if commentThread.Status == "" {
			// create thread
			err := createCommentThread(pr.Resource.PullRequestID, goFile, commentMsg)
			if err != nil {
				return err
			}
		} else {
			// add comment
			err := addComment(pr.Resource.PullRequestID, commentThread, commentMsg)
			if err != nil {
				return err
			}
			// set thread active
			err = setCommentThreadStatus(pr.Resource.PullRequestID, commentThread, 1)
			if err != nil {
				return err
			}
		}
	}

	// vote
	if len(missingTestGoFiles) == 0 {
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

func ReviewImages(pr *PullRequest) error {
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
}
