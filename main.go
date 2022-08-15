package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Versions struct {
	RepoName   string   `bson:"repoName"`
	BranchInfo []Branch `bson:"branches"`
}

type Branch struct {
	Name                      string   `bson:"name"`
	PublishOriginalBranchName bool     `bson:"pubishOriginalBranchName"`
	Active                    bool     `bson:"active"`
	Aliases                   []string `bson:"aliases"`
	GitBranchName             string   `bson:"gitBranchName"`
	IsStableBranch            bool     `bson:"isStableBranch"`
	UrlAliases                []string `bson:"urlAliases"`
	UrlSlug                   string   `bson:"urlSlug"`
	VersionSelectorLabel      string   `bson:"versionSelectorLabel"`
	BuildsWithSnooty          bool     `bson:"buildsWithSnooty"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	var uri string
	if uri = os.Getenv("MONGODB_URI"); uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	session, err := client.StartSession()
	if err != nil {
		panic(err)
	}

	defer func() {
		session.EndSession(context.TODO())
	}()

	coll := client.Database("version_flippers").Collection("versions")
	getVersionInfo(client, session, coll)

}
