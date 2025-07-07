package api

import "github.com/sammanbajracharya/drift/internal/store"

type RepoHandler struct {
	repoStore store.RepoStore
}

func NewRepoHandler(repoStore store.RepoStore) *RepoHandler {
	return &RepoHandler{
		repoStore: repoStore,
	}
}
