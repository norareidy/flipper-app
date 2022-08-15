package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAddVersion(t *testing.T) {
	uri := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	coll := client.Database("version_flippers").Collection("versions")
	fmt.Println("Branches array before changes:")
	print(coll)
	_, got := coll.BulkWrite(context.TODO(), createWriteModel("docs-golang", "1.12", "1.11"))
	fmt.Println("Branches array after changes:")
	print(coll)
	var want error = nil

	if want != got {
		t.Error("Failed to add verison.")
	}
}

func TestAddDuplicateVersion(t *testing.T) {
	got := checkVersionNumbers("1.12", "1.12")

	if got != "Warning: your old 'current' value 1.12 and new 'current' value 1.12 are the same.\nConfirm by typing '1.12' again, or type a new 'current' version number:" {
		t.Error("Didn't catch duplicate versions.")
	}

}

func TestAddDWeirdVersion(t *testing.T) {
	got := checkVersionNumbers("5.3", "1.12")

	if got != "Warning: your old 'current' value 1.12 and new 'current' value 5.30 seem far apart.\nConfirm by typing '5.30' again, or type a new 'current' version number:" {
		t.Error("Didn't catch abnormal version number.")
	}
}

func TestCorrectChangeNum(t *testing.T) {
	uri := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	result, err := client.Database("version_flippers").Collection("versions").BulkWrite(context.TODO(), createWriteModel("docs-golang", "1.15", "1.14"))
	got := result.ModifiedCount
	var want int64 = 3
	if got > want {
		t.Error("Too many branch elements changed.")
	} else if got < want {
		t.Error("Not all necessary changes were made.")
	}

}

func print(coll *mongo.Collection) {
	var branches []bson.D
	result, err := coll.Find(context.TODO(), bson.D{{"repoName", "docs-golang"}}, options.Find().SetProjection(bson.D{{"branches", 1}, {"_id", 0}}))
	if err != nil {
		panic(err)
	}
	if err = result.All(context.TODO(), &branches); err != nil {
		panic(err)
	}

	output, err := json.MarshalIndent(branches, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", output)
}
