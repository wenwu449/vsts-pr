package vsts

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/wenwu449/vsts-pr/ext"
)

type imageReview struct {
	diffs       *diffs
	pullRequest *PullRequest
}

func (r *imageReview) getBotCommentPrefix() string {
	return "[BOT_Image]\n"
}

func (r *imageReview) getBotCommentSuffix() string {
	return "\n*This comment was added by bot, please let me know if you have any suggestion!*"
}

func (r *imageReview) getFailedSign() string {
	signs := []string{":exclamation:", ":warning:", ":scream:", ":broken_heart:", ":no_entry:", ":bangbang:"}
	return signs[time.Now().Minute()%len(signs)]
}

func (r *imageReview) getPassedSign() string {
	signs := []string{":heavy_check_mark:", ":thumbsup:", ":white_check_mark:", ":ballot_box_with_check:", ":clap:", ":smiley:"}
	return signs[time.Now().Minute()%len(signs)]
}

func (r *imageReview) getPassedWord() string {
	words := []string{"Verified!", "Awesome!", "Clear!", "Great Job!"}
	return words[time.Now().Minute()%len(words)]
}

func (r *imageReview) getCommentContent(missingImages []string) (string, string) {
	essentialMessage := "All images listed on [acihealth](http://acihealth.azurewebsites.net/#cloudshell) are included in this list."
	if len(missingImages) == 0 {
		return essentialMessage, fmt.Sprintf(
			"%s%s %s %s\n%s",
			r.getBotCommentPrefix(),
			r.getPassedSign(),
			r.getPassedWord(),
			essentialMessage,
			r.getBotCommentSuffix())
	}
	sort.Strings(missingImages)
	essentialMessage = fmt.Sprintf("Following images should be included:**%+v**", missingImages)
	return essentialMessage, fmt.Sprintf(
		"%s%s %s\nYou can get image list from [acihealth](http://acihealth.azurewebsites.net/#cloudshell).\n%s",
		r.getBotCommentPrefix(),
		r.getFailedSign(),
		essentialMessage,
		r.getBotCommentSuffix())
}

func (r *imageReview) review() (bool, error) {
	log.Println("image check started.")

	var changedImageConfigs []imageConfig
	for _, imageConfig := range config.ImageConfigs {
		for _, change := range r.diffs.Changes {
			if strings.EqualFold(imageConfig.ConfigPath, change.Item.Path) {
				changedImageConfigs = append(changedImageConfigs, imageConfig)
				break
			}
		}
	}

	if len(changedImageConfigs) == 0 {
		log.Println("No change in image config")
		return true, nil
	}

	imageDistinct := make(map[string]map[string]struct{})
	for _, imageConfig := range config.ImageConfigs {
		imageDistinct[imageConfig.Os] = make(map[string]struct{})
	}

	done := make(chan http.Header)
	for _, endpoint := range config.Endpoints {
		go ext.GetHeaderFromHealthCheck(done, endpoint)
	}

	for range config.Endpoints {
		header := <-done
		if header != nil {
			for _, imageConfig := range config.ImageConfigs {
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
		err := getBranchItemContent(getBranchNameFromRefName(r.pullRequest.Resource.SourceRefName), imageConfig.ConfigPath, &imageList)
		if err != nil {
			return false, err
		}

		log.Printf("Checking: %s\n", imageConfig.ConfigPath)
		log.Printf("images 'commonImages': %+v\n", imageList.CommonImages)
		commonImages := []string{}
		if imageList.CommonImages != nil {
			for _, image := range imageList.CommonImages {
				commonImages = append(commonImages, image.Name)
			}
		} else {
			log.Printf("support legacy image config format: %+v\n", config.SupportLegacyImageFormat)
			if config.SupportLegacyImageFormat {
				log.Printf("images 'common': %+v\n", imageList.Common)
				commonImages = imageList.Common
			}
		}

		requiredImages := imageDistinct[imageConfig.Os]
		for imageVersion := range requiredImages {
			log.Printf("Checking required image: %s\n", imageVersion)
			found := false
			for _, image := range commonImages {
				version := image[strings.LastIndex(image, ":")+1:]
				if strings.EqualFold(version, imageVersion) {
					found = true
					break
				}
			}
			if !found {
				images = append(images, imageVersion)
				log.Printf("%s missing %s\n", imageConfig.ConfigPath, imageVersion)
			}
		}
		if len(images) > 0 {
			missingImagesMap[imageConfig.ConfigPath] = images
		}
	}

	if len(missingImagesMap) == 0 {
		log.Printf("image check passed.\n")
	} else {
		log.Printf("image check failed: %+v\n", missingImagesMap)
	}

	commentThreads, err := getCommentThreads(r.pullRequest.Resource.PullRequestID)
	if err != nil {
		return false, err
	}

	for _, imageConfig := range changedImageConfigs {
		missingImages, ok := missingImagesMap[imageConfig.ConfigPath]
		if !ok {
			missingImages = []string{}
		}

		commentThread := commentThread{}
		for _, thread := range commentThreads.Value {
			if !thread.IsDeleted && strings.EqualFold(thread.ThreadContext.FilePath, imageConfig.ConfigPath) {
				for _, comment := range thread.Comments {
					if comment.ID == 1 && comment.Author.ID == config.UserID && strings.HasPrefix(comment.Content, r.getBotCommentPrefix()) {
						commentThread = thread
						break
					}
				}
			}
		}

		essentialMessage, commentContent := r.getCommentContent(missingImages)

		status := 1
		if len(missingImages) == 0 {
			status = 2
		}

		if commentThread.Status == "" {
			// create thread
			err := createCommentThread(r.pullRequest.Resource.PullRequestID, imageConfig.ConfigPath, status, commentContent)
			if err != nil {
				return false, err
			}
		} else {
			// add comment
			err := addComment(r.pullRequest.Resource.PullRequestID, commentThread, essentialMessage, commentContent)
			if err != nil {
				return false, err
			}
			// set thread active
			err = setCommentThreadStatus(r.pullRequest.Resource.PullRequestID, commentThread, status)
			if err != nil {
				return false, err
			}
		}
	}

	log.Println("image check completed.")

	// review result
	return len(missingImagesMap) == 0, nil
}
