package main

import (
	"context"

	"github.com/google/go-github/v43/github"
)

var client = github.NewClient(nil)

func GetRespository(owner, repo string) (*github.Repository, error) {
	rep, _, err := client.Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		return nil, err
	}

	return rep, nil
}

func GetLatestRelease(owner, repo string) (*github.RepositoryRelease, error) {
	latestRelease, _, err := client.Repositories.GetLatestRelease(context.Background(), owner, repo)
	if err != nil {
		return nil, err
	}

	return latestRelease, nil
}

func GetReleaseAssets(owner, repo string, id int64) ([]*github.ReleaseAsset, error) {
	opts := &github.ListOptions{PerPage: 100}
	assets, _, err := client.Repositories.ListReleaseAssets(context.Background(), owner, repo, id, opts)
	if err != nil {
		return nil, err
	}

	return assets, nil
}
