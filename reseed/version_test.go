package reseed

import (
	"encoding/json"
	"net/http"
	"testing"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

func TestVersionActuallyChanged(t *testing.T) {
	// First, use the github API to get the latest github release
	resp, err := http.Get("https://api.github.com/repos/go-i2p/reseed-tools/releases/latest")
	if err != nil {
		t.Skipf("Failed to fetch GitHub release: %v", err)
	}
	defer resp.Body.Close()

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		t.Skipf("Failed to decode GitHub response: %v", err)
	}

	githubVersion := release.TagName
	if githubVersion == "" {
		t.Skip("No GitHub release found")
	}

	// Remove 'v' prefix if present
	if len(githubVersion) > 0 && githubVersion[0] == 'v' {
		githubVersion = githubVersion[1:]
	}

	// Next, compare it to the current version
	if Version == githubVersion {
		t.Fatal("Version not updated")
	}

	// Make sure the current version is larger than the previous version
	if Version < githubVersion {
		t.Fatalf("Version not incremented: current %s < github %s", Version, githubVersion)
	}
}
