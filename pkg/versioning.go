package pkg

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/cli/go-gh/v2/pkg/api"
)

var actionRefPattern = regexp.MustCompile(`[\w.-]+/[\w.-]+@latest`)

type release struct {
	TagName string `json:"tag_name"`
}

type commit struct {
	SHA string `json:"sha"`
}

// ResolveVersions scans content for `owner/repo@latest` placeholders and
// replaces each with a concrete version per mode: a semver tag
// (owner/repo@v1.2.3) or a pinned commit hash with an annotating comment
// (owner/repo@<sha> # v1.2.3). Unique action references are resolved
// concurrently so lookups don't block on each other.
func ResolveVersions(content string, mode Versioning) (string, error) {
	repos := uniqueRepos(content)
	if len(repos) == 0 {
		return content, nil
	}

	client, err := api.DefaultRESTClient()
	if err != nil {
		return "", fmt.Errorf("ghapi: %w", err)
	}

	type resolved struct {
		repo string
		ref  string
		err  error
	}

	results := make(chan resolved, len(repos))
	var wg sync.WaitGroup
	for _, repo := range repos {
		wg.Add(1)
		go func(repo string) {
			defer wg.Done()
			ref, err := resolveRef(client, repo, mode)
			results <- resolved{repo: repo, ref: ref, err: err}
		}(repo)
	}
	wg.Wait()
	close(results)

	out := content
	for r := range results {
		if r.err != nil {
			return "", fmt.Errorf("ghapi: resolving %s: %w", r.repo, r.err)
		}
		out = strings.ReplaceAll(out, r.repo+"@latest", r.ref)
	}
	return out, nil
}

func uniqueRepos(content string) []string {
	seen := make(map[string]struct{})
	var repos []string
	for _, match := range actionRefPattern.FindAllString(content, -1) {
		repo := strings.TrimSuffix(match, "@latest")
		if _, ok := seen[repo]; !ok {
			seen[repo] = struct{}{}
			repos = append(repos, repo)
		}
	}
	return repos
}

func resolveRef(client *api.RESTClient, repo string, mode Versioning) (string, error) {
	var rel release
	if err := client.Get(fmt.Sprintf("repos/%s/releases/latest", repo), &rel); err != nil {
		return "", err
	}

	if mode != VersioningHash {
		return fmt.Sprintf("%s@%s", repo, rel.TagName), nil
	}

	var c commit
	if err := client.Get(fmt.Sprintf("repos/%s/commits/%s", repo, rel.TagName), &c); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s@%s # %s", repo, c.SHA, rel.TagName), nil
}
