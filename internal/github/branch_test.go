package github

import (
	"fmt"
	"testing"

	"github.com/joho/godotenv"
)

func TestGetBranches(t *testing.T) {
	godotenv.Load("../../.env")

	gh, err := New()
	if err != nil {
		t.Fatalf("Could not create new GithubService: %e\n", err)
	}

	branches, err := gh.GetBranches("TobiasTheDanish", "go-track")
	if err != nil {
		t.Fatalf("Could not create new GithubService: %e\n", err)
	}

	fmt.Printf("branches %v\n", branches)
}
