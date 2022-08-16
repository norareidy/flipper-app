package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/DrSmithFr/go-console/pkg/color"
	"github.com/DrSmithFr/go-console/pkg/formatter"
	"github.com/DrSmithFr/go-console/pkg/style"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var io = style.NewConsoleStyler()

func previewAndUpdate(repo string, current string, oldCurrent string, sctx mongo.SessionContext, session mongo.Session, coll *mongo.Collection) {
	_, err := coll.BulkWrite(sctx, createWriteModel(repo, current, oldCurrent))
	if err != nil {
		panic(err)
	}
	previewChanges(getFullBranchArray(repo, sctx, coll), repo)
	commit := askToCommit()
	abortOrCommit(commit, session, sctx)
}

func createWriteModel(repo string, current string, oldCurrent string) []mongo.WriteModel {
	return []mongo.WriteModel{oldCurrentUpdates(repo, oldCurrent), newCurrentUpdates(repo, current), masterUpdates(repo)}

}

func oldCurrentUpdates(repo string, oldCurrent string) *mongo.UpdateOneModel {
	return mongo.NewUpdateOneModel().SetFilter(bson.D{{"repoName", repo}, {"branches.urlSlug", "current"}}).
		SetUpdate(bson.D{{"$set", bson.D{{"branches.$.urlSlug", ""}, {"branches.$.aliases", []string{""}},
			{"branches.$.urlAliases", []string{""}}, {"branches.$.versionSelectorLabel", oldCurrent},
			{"branches.$.isStableBranch", false}}}})
}

func newCurrentUpdates(repo string, current string) *mongo.UpdateOneModel {
	return mongo.NewUpdateOneModel().SetFilter(
		bson.D{{"repoName", repo}, {"branches.name", "master"}}).
		SetUpdate(bson.D{{"$set", bson.D{{"branches.$.name", "v" + current},
			{"branches.$.urlSlug", "current"}, {"branches.$.aliases", []string{"current"}},
			{"branches.$.gitBranchName", "v" + current}, {"branches.$.isStableBranch", true},
			{"branches.$.urlAliases", []string{"current"}}, {"branches.$.versionSelectorLabel", "v" + current + " (current)"}}}})
}

func masterUpdates(repo string) *mongo.UpdateOneModel {
	masterDoc := []interface{}{
		bson.D{{Key: "name", Value: "master"}, {Key: "publishOriginalBranchName", Value: true}, {Key: "active", Value: true},
			{Key: "aliases", Value: []string{"upcoming"}}, {Key: "gitBranchName", Value: "master"}, {Key: "isStableBranch", Value: false},
			{Key: "urlAliases", Value: []string{"upcoming"}}, {Key: "urlSlug", Value: "upcoming"}, {Key: "versionSelectorLabel", Value: "upcoming"},
			{Key: "buildsWithSnooty", Value: true}},
	}
	return mongo.NewUpdateOneModel().SetFilter(bson.D{{"repoName", repo}}).
		SetUpdate(bson.M{"$push": bson.M{"branches": bson.M{"$each": masterDoc, "$position": 0}}})

}

func previewChanges(branches []bson.D, repo string) error {
	io.Section("Preparing Changes")
	fmt.Printf("\nYour new version branch will be added to repo: %s", repo)
	fmt.Println("\n\nThe branches array will look like this: ")

	output, err := json.MarshalIndent(branches, "", "    ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", output)
	return nil
}

func getFullBranchArray(repo string, sctx mongo.SessionContext, coll *mongo.Collection) []bson.D {
	var branches []bson.D
	res, err := coll.Find(sctx, bson.D{{"repoName", repo}}, options.Find().SetProjection(bson.D{{"branches", 1}, {"_id", 0}}))
	if err != nil {
		panic(err)
	}
	if err = res.All(sctx, &branches); err != nil {
		panic(err)
	}
	return branches
}

func askToCommit() string {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("\nWould you like to commit these changes? Type 'changes are correct' to commit, or type 'reverse' to reverse: ")
	scanner.Scan()
	commit := scanner.Text()

	return commit

}

func abortOrCommit(commit string, session mongo.Session, sctx mongo.SessionContext) {
	if commit == "reverse" {
		s1 := formatter.NewOutputFormatterStyle(color.WHITE, color.BLUE, nil)
		fmt.Println(s1.Apply("\nOkay, changes were discarded.\n"))
		session.AbortTransaction(sctx)
	} else if commit == "changes are correct" {
		io.Success("Your new version branch was added.")
		session.CommitTransaction(sctx)
	} else {
		io.Error("Couldn't read input. Try again.")
		abortOrCommit(askToCommit(), session, sctx)
	}
}
