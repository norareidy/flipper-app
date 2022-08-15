# Version Flipper App

## Environment Variable

You must define your Go environment variable with your MongoDB connection string before using this app.

To do this, paste the following into the .env file at the root of your project:

`MONGODB_URI=mongodb+srv://<user>:<password>@<cluster-url>?retryWrites=true&w=majority`

Replace the connection string with your own MongoDB connection string. Visit the [MongoDB Docs](https://www.mongodb.com/docs/atlas/driver-connection/)
to learn more about connection strings.
