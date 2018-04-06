package vsts

import (
	"strings"
)

func getBranchNameFromRefName(refName string) string {
	return (strings.SplitAfterN(refName, "/", 3))[2]
}

func getdiffsURL(baseBranch string, targetBranch string) string {
	diffsURLTemplate := "https://{instance}/DefaultCollection/{project}/_apis/git/repositories/{repository}/diffs/commits?api-version={version}&targetVersionType=branch&targetVersion={targetBranch}&baseVersionType=branch&baseVersion={baseBranch}"
	r := strings.NewReplacer(
		"{instance}", config.Instance,
		"{project}", config.Project,
		"{repository}", config.Repo,
		"{version}", "1.0",
		"{baseBranch}", baseBranch,
		"{targetBranch}", targetBranch)

	return r.Replace(diffsURLTemplate)
}

func getDiffsBetweenBranches(baseBranch string, targetBranch string) (*diffs, error) {
	diffs := new(diffs)

	url := getdiffsURL(baseBranch, targetBranch)

	err := getFromVsts(url, diffs)
	if err != nil {
		return nil, err
	}

	return diffs, nil
}
