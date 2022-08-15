package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getVersionInfo(client *mongo.Client, session mongo.Session, coll *mongo.Collection) {

	err := mongo.WithSession(context.TODO(), session, func(sctx mongo.SessionContext) error {
		sctx.StartTransaction()
		defer func() {
			session.AbortTransaction(sctx)
		}()

		fmt.Println("Enter your repo name. Repo options:")
		repo := repoOptions(coll, sctx)

		var current string
		fmt.Print("Enter what you'd like the new 'current' version number to be\n\t- '1.7', for example: ")
		fmt.Scan(&current)

		var version []Versions
		cursor, err := coll.Find(sctx, bson.D{{"repoName", repo}})
		if err = cursor.All(sctx, &version); err != nil {
			panic(err)
		}

		oldCurrent := version[0].BranchInfo[1].Name
		fmt.Print(checkVersionNumbers(current, oldCurrent))
		fmt.Scan(&current)

		_, err = coll.BulkWrite(sctx, createWriteModel(repo, current, oldCurrent))
		if err = printInfo(repo, sctx, coll); err != nil {
			panic(err)
		}
		askToCommit(session, sctx)
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func repoOptions(coll *mongo.Collection, sctx mongo.SessionContext) string {
	var repo string
	var allVersions []Versions
	var allRepos []string
	cursor, err := coll.Find(sctx, bson.D{{}})
	if err = cursor.All(sctx, &allVersions); err != nil {
		panic(err)
	}
	for _, r := range allVersions {
		fmt.Println(r.RepoName)
		allRepos = append(allRepos, r.RepoName)
	}
	fmt.Print("Type your repo choice: ")
	fmt.Scan(&repo)

	for _, r := range allRepos {
		if r == repo {
			return repo
		}
	}
	fmt.Println("ERROR: your input is not a repo option. Try again.\n---------------------------------------")
	return repoOptions(coll, sctx)
}

func createWriteModel(repo string, current string, oldCurrent string) []mongo.WriteModel {
	masterDoc := []interface{}{
		bson.D{{Key: "name", Value: "master"}, {Key: "publishOriginalBranchName", Value: true}, {Key: "active", Value: true},
			{Key: "aliases", Value: []string{"upcoming"}}, {Key: "gitBranchName", Value: "master"}, {Key: "isStableBranch", Value: false},
			{Key: "urlAliases", Value: []string{"upcoming"}}, {Key: "urlSlug", Value: "upcoming"}, {Key: "versionSelectorLabel", Value: "upcoming"},
			{Key: "buildsWithSnooty", Value: true}},
	}

	models := []mongo.WriteModel{
		mongo.NewUpdateOneModel().SetFilter(
			bson.D{{"repoName", repo}, {"branches.urlSlug", "current"}}).
			SetUpdate(bson.D{{"$set", bson.D{{"branches.$.urlSlug", ""}, {"branches.$.aliases", []string{""}},
				{"branches.$.urlAliases", []string{""}}, {"branches.$.versionSelectorLabel", oldCurrent},
				{"branches.$.isStableBranch", false}}}}),
		mongo.NewUpdateOneModel().SetFilter(
			bson.D{{"repoName", repo}, {"branches.name", "master"}}).
			SetUpdate(bson.D{{"$set", bson.D{{"branches.$.name", "v" + current},
				{"branches.$.urlSlug", "current"}, {"branches.$.aliases", []string{"current"}},
				{"branches.$.gitBranchName", "v" + current}, {"branches.$.isStableBranch", true},
				{"branches.$.urlAliases", []string{"current"}}, {"branches.$.versionSelectorLabel", "v" + current + " (current)"}}}}),
		mongo.NewUpdateOneModel().SetFilter(bson.D{{"repoName", repo}}).
			SetUpdate(bson.M{"$push": bson.M{"branches": bson.M{"$each": masterDoc, "$position": 0}}}),
	}
	return models

}

func checkVersionNumbers(c1 string, c2 string) string {
	current, err := strconv.ParseFloat(c1, 32)
	if err != nil {
		panic(err)
	}
	oldCurrent, err := strconv.ParseFloat(strings.TrimPrefix(c2, "v"), 32)

	if oldCurrent > current && (math.Mod(current, 1) != 0) {
		return fmt.Sprintf("Warning: your old 'current' value %.2f seems to be greater than your new 'current' value %s.\nConfirm by typing '%s' again, or type a new 'current' version number:",
			oldCurrent, c1, c1)
	}
	if math.Abs(oldCurrent-current) > 0.3 {
		return fmt.Sprintf("Warning: your old 'current' value %.2f and new 'current' value %s seem far apart.\nConfirm by typing '%s' again, or type a new 'current' version number:",
			oldCurrent, c1, c1)
	}
	if oldCurrent == current {
		return fmt.Sprintf("Warning: your old 'current' value %.2f and new 'current' value %s are the same.\nConfirm by typing '%s' again, or type a new 'current' version number:",
			oldCurrent, c1, c1)
	}
	return fmt.Sprintf("New current version number %s passes validation, but you can still change it.\nConfirm by typing '%s' again, or type a new 'current' version number:", c1, c1)
}

func printInfo(repo string, sctx mongo.SessionContext, coll *mongo.Collection) error {
	fmt.Println("\n-----------------------------------------------------------")
	fmt.Printf("\nYour new version branch will be added to repo: %s", repo)
	fmt.Println("\n\nThe branches array will look like this: ")
	var branches []bson.D
	res, err := coll.Find(sctx, bson.D{{"repoName", repo}}, options.Find().SetProjection(bson.D{{"branches", 1}, {"_id", 0}}))
	if err != nil {
		fmt.Println("Error, someone else may have a transaction open.")
		return err
	}
	if err = res.All(sctx, &branches); err != nil {
		return err
	}

	output, err := json.MarshalIndent(branches, "", "    ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", output)
	return nil
}

func askToCommit(session mongo.Session, sctx mongo.SessionContext) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("\nWould you like to commit these changes? Type 'changes are correct' to commit, or type 'reverse' to reverse: ")
	scanner.Scan()
	undo := scanner.Text()

	if undo == "reverse" {
		fmt.Println("Okay, changes were discarded.")
		session.AbortTransaction(sctx)
	} else if undo == "changes are correct" {
		fmt.Println("Great, your new version branch was added.")
		session.CommitTransaction(sctx)
	} else {
		fmt.Println("Error, couldn't read input. Try again.")
		askToCommit(session, sctx)
	}
}
