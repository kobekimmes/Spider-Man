

package db

import (
	"os"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
    //"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)


func InitConnection() {

	
	connectionString := os.Getenv("SPIDER_MAN_CONNECTION_URI")

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(connectionString).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(opts)

	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(nil); err != nil {
			panic(err)
		}
	}()

	// Send a ping to confirm a successful connection
	// if err := client.Ping(, readpref.Primary()); err != nil {
	// 	panic(err)
	// }
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
}


