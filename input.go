package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/DrSmithFr/go-console/pkg/color"
	"github.com/DrSmithFr/go-console/pkg/formatter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func collectInputAndValidate(sctx mongo.SessionContext, coll *mongo.Collection) (string, string, string) {
	fmt.Println("Enter your repo name. Repo options:")
	allRepos := repoOptions(coll, sctx)
	io.Listing(repoOptions(coll, sctx))

	repo := getRepoInput()
	repo = validateRepo(allRepos, repo)

	current := getCurrentInput()
	oldCurrent := getOldCurrent(coll, sctx, repo)
	current = validateCurrentValues(current, oldCurrent)

	return repo, current, oldCurrent
}

func repoOptions(coll *mongo.Collection, sctx mongo.SessionContext) []string {
	var allVersions []Repo
	var allRepos []string

	cursor, err := coll.Find(sctx, bson.D{{}})
	if err != nil {
		panic(err)
	}

	if err = cursor.All(sctx, &allVersions); err != nil {
		panic(err)
	}
	for _, r := range allVersions {
		allRepos = append(allRepos, r.RepoName)
	}

	return allRepos

}

func getRepoInput() string {
	var repo string
	fmt.Print("Type your repo name: ")
	fmt.Scan(&repo)
	return repo
}

func validateRepo(allRepos []string, repo string) string {
	if !checkRepoOptions(allRepos, repo) {
		io.Warning("Your input is not a repo option. Try again.")
		return getRepoInput()
	}
	return repo
}

func checkRepoOptions(allRepos []string, repo string) bool {
	for _, r := range allRepos {
		if r == repo {
			return true
		}
	}
	return false
}

func getCurrentInput() string {
	var current string
	fmt.Print("Enter what you'd like the new 'current' version number to be\n\t- '1.7', for example: ")
	fmt.Scan(&current)
	return current
}

func getOldCurrent(coll *mongo.Collection, sctx mongo.SessionContext, repo string) string {
	var version Repo
	err := coll.FindOne(sctx, bson.D{{"repoName", repo}}).Decode(&version)
	if err != nil {
		panic(err)
	}

	oldCurrent := version.BranchInfo[1].Name
	return oldCurrent
}

func validateCurrentValues(current string, oldCurrent string) string {
	if !isNumber(current) {
		io.Caution("Your 'current' value " + current + " is not a number.")
		return currentRedo()
	} else if !isNumber(strings.TrimPrefix(oldCurrent, "v")) {
		io.Caution("Can't validate input.")
		return currentRedo()
	}

	if areVersionsCorrect(current, strings.TrimPrefix(oldCurrent, "v")) {
		s1 := formatter.NewOutputFormatterStyle(color.BLACK, color.GREEN, nil)
		fmt.Println(s1.Apply("New current version number " + current + " passes validation."))
	} else {
		return currentRedo()
	}
	return current
}

func isNumber(parseString string) bool {
	_, err := strconv.ParseFloat(parseString, 32)
	return err == nil
}

func areVersionsCorrect(current string, oldCurrent string) bool {
	currNum, _ := strconv.ParseFloat(current, 32)
	oldCurrNum, _ := strconv.ParseFloat(oldCurrent, 32)

	if oldCurrNum > currNum {
		io.Caution("Your old 'current' value " + oldCurrent + " seems to be greater than your new 'current' value " + current + ".")
		return false
	} else if math.Abs(oldCurrNum-currNum) > 0.3 {
		io.Caution("Your old 'current' value " + oldCurrent + " and new 'current' value " + current + " seem far apart.")
		return false
	} else if oldCurrNum == currNum {
		io.Caution("Your old 'current' value " + oldCurrent + " and new 'current' value " + current + " are the same.")
		return false
	}
	return true
}

func currentRedo() string {
	var curr string
	fmt.Print("Confirm your 'current' input by typing it again, or type a new 'current' version number: ")
	fmt.Scan(&curr)
	return curr
}
