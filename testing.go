package main

/*import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
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

	db := client.Database("myDB")
	opts := options.GridFSBucket().SetName("custom name")
	bucket, err := gridfs.NewBucket(db, opts)

	if err != nil {
		panic(err)
	}

	fileContent, err := os.Open("path/to/file.txt")
	uploadOpts := options.GridFSUpload().SetMetadata(bson.D{{"metadata tag", "first"}})

	objectID, err := bucket.UploadFromStream("file.txt", io.Reader(fileContent),
		uploadOpts)
	if err != nil {
		panic(err)
	}

	fmt.Printf("New file uploaded with ID %s", objectID)

}*/
