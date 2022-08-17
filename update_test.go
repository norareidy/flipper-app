package main

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestWrite(t *testing.T) {
	originalDocs := getOriginal()
	testColl := mongoSetup()
	_, err := testColl.InsertOne(context.TODO(), originalDocs)
	if err != nil {
		panic(err)
	}
	_, err = testColl.BulkWrite(context.TODO(), createWriteModel("docs-golang", "1.11", "v1.10"))
	if err != nil {
		panic(err)
	}
	var got Repo
	err = testColl.FindOne(context.TODO(), bson.D{{"repoName", "docs-golang"}}).Decode(&got)
	if err != nil {
		panic(err)
	}
	testColl.Drop(context.TODO())
	resultDocs := getResult()
	if !reflect.DeepEqual(got, resultDocs) {
		t.Error("Collection documents were not updated correctly.")
	}
}

func TestPreviewChanges(t *testing.T) {
	master := bson.D{{"name", "master"}, {"urlSlug", "upcoming"}}
	current := bson.D{{"name", "v1.11"}, {"urlSlug", "current"}}
	oldCurrent := bson.D{{"name", "v1.10"}, {"urlSlug", ""}}

	branches := []bson.D{
		master, current, oldCurrent,
	}
	repo := "docs-golang"
	got := previewChanges(branches, repo)

	if got != nil {
		t.Error("Failed to preview changes.")
	}
}

func getOriginal() Repo {
	var firstMaster = Branch{Name: "master", PublishOriginalBranchName: true, Active: true, Aliases: []string{"upcoming"},
		GitBranchName: "master", IsStableBranch: false, UrlAliases: []string{"upcoming"}, UrlSlug: "upcoming",
		VersionSelectorLabel: "upcoming", BuildsWithSnooty: true}

	var firstCurrent = Branch{Name: "v1.10", PublishOriginalBranchName: true, Active: true, Aliases: []string{"current"},
		GitBranchName: "v1.10", IsStableBranch: true, UrlAliases: []string{"current"}, UrlSlug: "current",
		VersionSelectorLabel: "v1.10 (current)", BuildsWithSnooty: true}

	var firstOldCurrent = Branch{Name: "v1.9", PublishOriginalBranchName: true, Active: true, Aliases: []string{""},
		GitBranchName: "v1.9", IsStableBranch: false, UrlAliases: []string{""}, UrlSlug: "",
		VersionSelectorLabel: "v1.9", BuildsWithSnooty: true}

	var firstGolang = Repo{RepoName: "docs-golang", BranchInfo: []Branch{firstMaster, firstCurrent, firstOldCurrent}}
	return firstGolang
}

func getResult() Repo {
	var resultMaster = Branch{Name: "master", PublishOriginalBranchName: false, Active: true, Aliases: []string{"upcoming"},
		GitBranchName: "master", IsStableBranch: false, UrlAliases: []string{"upcoming"}, UrlSlug: "upcoming",
		VersionSelectorLabel: "upcoming", BuildsWithSnooty: true}

	var resultCurrent = Branch{Name: "v1.11", PublishOriginalBranchName: true, Active: true, Aliases: []string{"current"},
		GitBranchName: "v1.11", IsStableBranch: true, UrlAliases: []string{"current"}, UrlSlug: "current",
		VersionSelectorLabel: "v1.11 (current)", BuildsWithSnooty: true}

	var resultOldCurrent = Branch{Name: "v1.10", PublishOriginalBranchName: true, Active: true, Aliases: []string{""},
		GitBranchName: "v1.10", IsStableBranch: false, UrlAliases: []string{""}, UrlSlug: "",
		VersionSelectorLabel: "v1.10", BuildsWithSnooty: true}

	var resultOlder = Branch{Name: "v1.9", PublishOriginalBranchName: true, Active: true, Aliases: []string{""},
		GitBranchName: "v1.9", IsStableBranch: false, UrlAliases: []string{""}, UrlSlug: "",
		VersionSelectorLabel: "v1.9", BuildsWithSnooty: true}

	var resultGolang = Repo{RepoName: "docs-golang", BranchInfo: []Branch{resultMaster, resultCurrent, resultOldCurrent, resultOlder}}
	return resultGolang
}

func mongoSetup() *mongo.Collection {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	uri := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	testColl := client.Database("version_flippers").Collection("test")

	return testColl
}
