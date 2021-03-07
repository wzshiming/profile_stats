package source

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	githubv3 "github.com/google/go-github/v33/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Stat struct {
	Name          string
	Stars         int
	Forks         int
	Issues        int
	Commits       int
	Reviews       int
	PullRequests  int
	ContributedTo int
}

func NewSource(token string) *Source {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return &Source{
		token: token,
		cliv3: githubv3.NewClient(httpClient),
		cliv4: githubv4.NewClient(httpClient),
	}
}

type Source struct {
	token string
	cliv3 *githubv3.Client
	cliv4 *githubv4.Client
}

func (s *Source) Stat(ctc context.Context, username string) (*Stat, error) {
	var query struct {
		User struct {
			Repositories struct {
				Nodes []struct {
					StargazerCount githubv4.Int
					ForkCount      githubv4.Int
					IsFork         githubv4.Boolean
				}
			} `graphql:"repositories(first: 100, ownerAffiliations: OWNER, orderBy: {direction: DESC, field: STARGAZERS})"`
			Contributions struct {
				TotalCommitContributions            githubv4.Int
				TotalPullRequestReviewContributions githubv4.Int
			} `graphql:"contributionsCollection"`
			ContributedTo struct {
				TotalCount githubv4.Int
			} `graphql:"repositoriesContributedTo(first: 0)"`
			PullRequests struct {
				TotalCount githubv4.Int
			} `graphql:"pullRequests(first: 0)"`
			Issues struct {
				TotalCount githubv4.Int
			} `graphql:"issues(first: 0)"`
			// Login githubv4.String
			Name githubv4.String
		} `graphql:"user(login: $username)"`
	}

	variables := map[string]interface{}{
		"username": githubv4.String(username),
	}

	err := s.cliv4.Query(ctc, &query, variables)
	if err != nil {
		return nil, err
	}
	stat := Stat{}
	stat.Name = string(query.User.Name)
	for _, repo := range query.User.Repositories.Nodes {
		stat.Stars += int(repo.StargazerCount)
		if !repo.IsFork {
			stat.Forks += int(repo.ForkCount)
		}
	}

	stat.Commits = int(query.User.Contributions.TotalCommitContributions)
	stat.Reviews = int(query.User.Contributions.TotalPullRequestReviewContributions)
	stat.ContributedTo = int(query.User.ContributedTo.TotalCount)
	stat.PullRequests = int(query.User.PullRequests.TotalCount)
	stat.Issues = int(query.User.Issues.TotalCount)
	return &stat, nil
}

func (s *Source) CommitCounter(ctx context.Context, username string) (int, error) {
	result, _, err := s.cliv3.Search.Commits(ctx, fmt.Sprintf("author:%q", username), &githubv3.SearchOptions{
		ListOptions: githubv3.ListOptions{PerPage: 1},
	})
	if err != nil {
		return 0, err
	}
	return *result.Total, nil
}

func (s *Source) UploadGist(ctx context.Context, owner, description, name string, r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	dataContext := string(data)

	var oriGist *githubv3.Gist
	_, err = s.listGist(ctx, owner, func(gists []*githubv3.Gist) bool {
		for _, gist := range gists {
			if gist.Description != nil && *gist.Description == description {
				oriGist = gist
				return false
			}
		}
		return true
	})
	if err != nil {
		return "", err
	}

	var raw string
	if oriGist == nil {
		gist, _, err := s.cliv3.Gists.Create(ctx, &githubv3.Gist{
			Public: githubv3.Bool(true),
			Files: map[githubv3.GistFilename]githubv3.GistFile{
				githubv3.GistFilename(name): {
					Content: &dataContext,
				},
			},
			Description: &description,
		})
		if err != nil {
			return "", err
		}
		raw = *gist.Files[githubv3.GistFilename(name)].RawURL
	} else {
		file := oriGist.Files[githubv3.GistFilename(name)]
		if file.Content != nil && *file.Content == dataContext {
			raw = *oriGist.Files[githubv3.GistFilename(name)].RawURL
		} else {
			oriGist.Files[githubv3.GistFilename(name)] = githubv3.GistFile{
				Filename: &name,
				Content:  &dataContext,
			}
			gist, _, err := s.cliv3.Gists.Edit(ctx, *oriGist.ID, oriGist)
			if err != nil {
				return "", err
			}
			raw = *gist.Files[githubv3.GistFilename(name)].RawURL
		}
	}
	raw = strings.SplitN(raw, "/raw/", 2)[0] + "/raw/" + name
	return raw, nil
}

func (s *Source) UploadAsset(ctx context.Context, owner, repo, release, name string, r io.Reader) (string, error) {
	var releaseID *int64
	_, err := s.listReleases(ctx, owner, repo, func(releases []*githubv3.RepositoryRelease) bool {
		for _, r := range releases {
			if r.Name != nil && *r.Name == release {
				releaseID = r.ID
				return false
			}
		}
		return true
	})
	if err != nil {
		return "", err
	}

	if releaseID == nil {
		respRelease, _, err := s.cliv3.Repositories.CreateRelease(ctx, owner, repo, &githubv3.RepositoryRelease{
			Name:    &release,
			TagName: &release,
		})
		if err != nil {
			return "", err
		}
		releaseID = respRelease.ID
	}

	dir, err := ioutil.TempDir(os.TempDir(), "profile_stats_asset")
	if err != nil {
		return "", err
	}

	filename := filepath.Join(dir, name)
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	if err != nil {
		return "", err
	}
	err = f.Sync()
	if err != nil {
		return "", err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return "", err
	}

	respAsset, _, err := s.cliv3.Repositories.UploadReleaseAsset(ctx, owner, repo, *releaseID, &githubv3.UploadOptions{
		Name: name,
	}, f)
	if err != nil {
		return "", err
	}
	return *respAsset.URL, nil
}

func (s *Source) listReleases(ctx context.Context, owner, repo string, next func([]*githubv3.RepositoryRelease) bool) ([]*githubv3.RepositoryRelease, error) {
	opt := &githubv3.ListOptions{
		PerPage: 100,
	}
	var out []*githubv3.RepositoryRelease
	for {
		list, resp, err := s.cliv3.Repositories.ListReleases(ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}
		if next != nil && !next(list) {
			return out, nil
		}
		out = append(out, list...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return out, nil
}

func (s *Source) listGist(ctx context.Context, owner string, next func([]*githubv3.Gist) bool) ([]*githubv3.Gist, error) {
	opt := githubv3.ListOptions{
		PerPage: 100,
	}
	var out []*githubv3.Gist
	for {
		list, resp, err := s.cliv3.Gists.List(ctx, owner, &githubv3.GistListOptions{
			ListOptions: opt,
		})
		if err != nil {
			return nil, err
		}
		if next != nil && !next(list) {
			return out, nil
		}
		out = append(out, list...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return out, nil
}

func (s *Source) UploadGit(ctx context.Context, owner, repo, branch, name string, r io.Reader) (string, error) {
	giturl := "https://github.com/" + owner + "/" + repo

	auth := &githttp.BasicAuth{
		Username: owner,
		Password: s.token,
	}

	dir, err := ioutil.TempDir(os.TempDir(), "profile_stats_git")
	if err != nil {
		return "", err
	}
	dir += dir + "/git"

	remoteName := "origin-" + branch
	refName := plumbing.NewBranchReferenceName(branch)
	fetch := []gitconfig.RefSpec{
		gitconfig.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%[1]s", branch, remoteName)),
	}

	var resp *git.Repository
	_, err = os.Stat(dir + "/.git")
	if err == nil {
		resp, err = git.PlainOpen(dir)
	} else {
		resp, err = git.PlainInit(dir, false)
	}
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, dir)
	}
	err = resp.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, refName))
	if err != nil {
		return "", err
	}

	remote, err := resp.Remote(remoteName)
	if err != nil {
		if err != git.ErrRemoteNotFound {
			return "", err
		}
		c := &gitconfig.RemoteConfig{
			Name:  remoteName,
			URLs:  []string{giturl},
			Fetch: fetch,
		}
		remote, err = resp.CreateRemote(c)
		if err != nil {
			return "", err
		}
	}

	_, err = resp.Branch(branch)
	if err != nil {
		if err != git.ErrBranchNotFound {
			return "", err
		}
		err = resp.CreateBranch(&gitconfig.Branch{
			Name:   branch,
			Merge:  refName,
			Remote: remoteName,
		})
		if err != nil {
			return "", err
		}
		_, err = resp.Branch(branch)
		if err != nil {
			return "", err
		}
	}

	err = remote.FetchContext(ctx, &git.FetchOptions{
		RemoteName: remoteName,
		RefSpecs:   fetch,
		Progress:   os.Stderr,
		Auth:       auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		if _, ok := err.(git.NoMatchingRefSpecError); !ok {
			return "", fmt.Errorf("git fetch: %w", err)
		}
	}

	refIter, err := resp.Storer.IterReferences()
	if err != nil {
		return "", fmt.Errorf("iterReferences: %w", err)
	}
	ref, err := refIter.Next()
	if err != nil {
		return "", fmt.Errorf("next: %w", err)
	}
	if !ref.Hash().IsZero() {
		err = resp.Storer.SetReference(plumbing.NewHashReference(refName, ref.Hash()))
		if err != nil {
			return "", fmt.Errorf("setReference: %w", err)
		}

		work, err := resp.Worktree()
		if err != nil {
			return "", err
		}
		err = work.Reset(&git.ResetOptions{
			Commit: ref.Hash(),
			Mode:   git.HardReset,
		})
		if err != nil {
			return "", fmt.Errorf("git reset: %w", err)
		}
	}

	fname := filepath.Join(dir, name)
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(f, r)
	if err != nil {
		f.Close()
		return "", err
	}
	f.Close()

	work, err := resp.Worktree()
	if err != nil {
		return "", err
	}
	_, err = work.Add(name)
	if err != nil {
		return "", fmt.Errorf("git add: %w", err)
	}
	status, err := work.Status()
	if err != nil {
		return "", err
	}

	if len(status) != 0 &&
		status[name] != nil &&
		(status[name].Staging != git.Unmodified || status[name].Worktree != git.Unmodified) {
		now := time.Now()
		_, err = work.Commit(fmt.Sprintf("Automatic update %s", now.Format(time.RFC3339)), &git.CommitOptions{
			Author: &object.Signature{
				Name: "bot",
				When: now,
			},
		})
		if err != nil {
			return "", fmt.Errorf("git commit: %w", err)
		}
		err = resp.PushContext(ctx, &git.PushOptions{
			Auth:       auth,
			RemoteName: remoteName,
			Progress:   os.Stderr,
		})
		if err != nil {
			return "", fmt.Errorf("git push: %w", err)
		}
	}

	return giturl + "/raw/" + branch + "/" + name, nil
}
