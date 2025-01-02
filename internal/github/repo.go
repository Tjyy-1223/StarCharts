package github

import (
	"context"
	"errors"
	"net/http"
)

type Repository struct {
	FullName        string `json:"full_name"`
	StargazersCount int    `json:"stargazers_count"`
	CreateAt        int    `json:"create_at"`
}

var ErrNotFound = errors.New("repository not found")

// RepoDetails gets the given repository details.
func (gh *GitHub) RepoDetails(ctx context.Context, name string) (Repository, error) {

}

func (gh *GitHub) makeRepoRequest(ctx context.Context, name, etag string) (*http.Response, error) {

}
