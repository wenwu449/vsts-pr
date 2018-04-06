package vsts

import (
	"strings"
)

func getBranchItemURL(branch string, itemPath string) string {
	itemURLTemplate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/items?api-version={version}&versionType={versionType}&version={versionValue}&scopePath={itemPath}&lastProcessedChange=true"
	r := strings.NewReplacer(
		"{instance}", config.Instance,
		"{project}", config.Project,
		"{repository}", config.Repo,
		"{versionType}", "branch",
		"{versionValue}", branch,
		"{itemPath}", itemPath,
		"{version}", "1.0")

	return r.Replace(itemURLTemplate)

}

func getBranchItemContent(branch string, itemPath string, v interface{}) error {

	url := getBranchItemURL(branch, itemPath)
	err := getFromVsts(url, v)

	if err != nil {
		return err
	}

	return nil
}
