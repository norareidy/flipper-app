package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
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
		panic("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
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
	//b := Branch{Name: "efef", Active: true}
	defer func() {
		session.EndSession(context.TODO())
	}()

	coll := client.Database("version_flippers").Collection("versions")

	err = mongo.WithSession(context.TODO(), session, func(sctx mongo.SessionContext) error {
		sctx.StartTransaction()
		defer func() {
			session.AbortTransaction(sctx)
		}()

		io.Title("Starting version updater .......")
		checkOtherTransactions(coll, sctx)

		repo, current, oldCurrent := collectInputAndValidate(sctx, coll)
		previewAndUpdate(repo, current, oldCurrent, sctx, session, coll)
		return nil
	})

	if err != nil {
		panic(err)
	}
}

func checkOtherTransactions(coll *mongo.Collection, sctx mongo.SessionContext) {
	var updatedDoc bson.D
	err := coll.FindOneAndUpdate(sctx, bson.D{{"repoName", "docs-golang"}}, bson.D{{"$set", bson.D{{"repoName", "docs-golang"}}}}).Decode(&updatedDoc)
	if err != nil {
		io.Error("Error, someone else may have a transaction open.")
		panic(err)
	}
}
