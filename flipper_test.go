package main

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAddVersion(t *testing.T) {
	//test code
	uri := "mongodb+srv://m001-student:m001-mongodb-basics@sandbox.csalm.mongodb.net/?retryWrites=true&w=majority"
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	_, got := client.Database("version_flippers").Collection("versions").BulkWrite(context.TODO(), updateVersions("docs-golang", "1.13", "1.12"))
	var want error = nil

	if want != got {
		t.Error("Failed to add verison.")
	}
}

func TestAddDuplicateVersion(t *testing.T) {
	//test code
	uri := "mongodb+srv://m001-student:m001-mongodb-basics@sandbox.csalm.mongodb.net/?retryWrites=true&w=majority"
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	_, got := client.Database("version_flippers").Collection("versions").BulkWrite(context.TODO(), updateVersions("docs-golang", "1.12", "1.12"))

	if got == nil {
		t.Error("Didn't catch duplicate versions.")
	}

}

func TestAddDWeirdVersion(t *testing.T) {
	uri := "mongodb+srv://m001-student:m001-mongodb-basics@sandbox.csalm.mongodb.net/?retryWrites=true&w=majority"
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	_, got := client.Database("version_flippers").Collection("versions").BulkWrite(context.TODO(), updateVersions("docs-golang", "4.2", "1.13"))
	if got == nil {
		t.Error("Didn't catch abnormal version number.")
	}
}

func TestCorrectChangeNum(t *testing.T) {
	uri := "mongodb+srv://m001-student:m001-mongodb-basics@sandbox.csalm.mongodb.net/?retryWrites=true&w=majority"
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	result, err := client.Database("version_flippers").Collection("versions").BulkWrite(context.TODO(), updateVersions("docs-golang", "1.15", "1.14"))
	got := result.ModifiedCount
	var want int64 = 3
	if got > want {
		t.Error("Too many branch elements changed.")
	} else if got < want {
		t.Error("Not all necessary changes were made.")
	}

}
