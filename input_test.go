package main

import (
	"testing"
)

func TestEnterDuplicateVersion(t *testing.T) {
	got := areVersionsCorrect("1.12", "1.12")

	if got != false {
		t.Error("Didn't catch duplicate versions.")
	}

}

func TestEnterIrregularVersion(t *testing.T) {
	got := areVersionsCorrect("5.3", "1.12")

	if got != false {
		t.Error("Didn't catch abnormal version number.")
	}
}

func TestEnterSmallVersion(t *testing.T) {
	got := areVersionsCorrect("4.7", "4.9")

	if got != false {
		t.Error("Didn't catch low version number.")
	}
}

func TestValidRepo(t *testing.T) {
	repos := []string{"docs-golang", "docs-node", "docs-java"}
	got := validateRepo(repos, "docs-node")

	if got != "docs-node" {
		t.Error("Error validating repo.")
	}
}

func TestInvalidRepo(t *testing.T) {
	repos := []string{"docs-golang", "docs-node", "docs-java"}
	got := checkRepoOptions(repos, "docs-go")

	if got != false {
		t.Error("Didn't catch invalid repo input.")
	}
}

func TestStringVersion(t *testing.T) {
	repoInput := "beta"
	got := isNumber(repoInput)

	if got != false {
		t.Error("Didn't catch string version input.")
	}
}
