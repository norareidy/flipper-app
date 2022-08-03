package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
	err = mongo.WithSession(context.TODO(), session, func(sctx mongo.SessionContext) error {
		sctx.StartTransaction()
		var repo string
		fmt.Print("Enter your repo name\n\t- 'docs-node', for example: ")
		fmt.Scan(&repo)

		var oldCurrent string
		fmt.Print("Enter the number of the old 'current' version \n\t- '1.6', for example: ")
		fmt.Scan(&oldCurrent)

		var current string
		fmt.Print("Now enter what you'd like the new 'current' version number to be\n\t- '1.7', for example: ")
		fmt.Scan(&current)

		coll := client.Database("version_flippers").Collection("versions")

		results, err := coll.BulkWrite(sctx, updateVersions(repo, current, oldCurrent))
		if err != nil {
			log.Fatal(err)
			return err
		}

		fmt.Println("\n-----------------------------------------------------------")
		fmt.Printf("\nYour new version branch was added!\n\tAdded to repo: %s \n\tNumber of versions changed: %d", repo, results.ModifiedCount)
		fmt.Println("\n\nBranch array elements changed: ")
		var branches []bson.D
		r, err := coll.Find(sctx, bson.D{{"repoName", repo}}, options.Find().SetProjection(bson.D{{"branches", 1}, {"_id", 0}}))
		if err = r.All(sctx, &branches); err != nil {
			return err
		}

		output, err := json.MarshalIndent(branches, "", "    ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", output)

		var undo string
		fmt.Print("\nWould you like to undo these changes? Type 'yes' or 'no': ")
		fmt.Scan(&undo)

		if undo == "yes" {
			session.AbortTransaction(sctx)
			fmt.Println("Okay, changes were discarded.")
		} else {
			session.CommitTransaction(sctx)
			fmt.Println("Great, process is complete.")
		}
		return nil
	})
	session.EndSession(context.TODO())

}

func updateVersions(repo string, current string, oldCurrent string) []mongo.WriteModel {
	first, err := strconv.ParseFloat(current, 32)
	if err != nil {
		log.Fatal("Please enter a numeric value.")
	}
	second, err := strconv.ParseFloat(oldCurrent, 32)
	if err != nil {
		log.Fatal("Please enter a numeric value.")
	}
	if (second == first) || math.Abs(second-first) > 0.5 {
		fmt.Println(second - first)
		log.Fatal("Version numbers are incorrect. Please check your versions and try again.")
	}

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
