package database

import (
	"context"
	"time"

	"github.com/qiniu/qmgo/options"
)

// MongoCreateIndex - create one index for a mongo collection
func MongoCreateIndex(dbName, collectionName string, index options.IndexModel) error {
	client := GetMongo()
	db := client.Database(dbName)               // set database name
	collection := db.Collection(collectionName) // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// create index
	return collection.CreateOneIndex(ctx, index)
}

// MongoCreateIndexes - create many indexes for a mongo collection
func MongoCreateIndexes(dbName, collectionName string, indexes []options.IndexModel) error {
	client := GetMongo()
	db := client.Database(dbName)               // set database name
	collection := db.Collection(collectionName) // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// create indexes
	return collection.CreateIndexes(ctx, indexes)
}

// MongoDropIndex - drop one/multiple indexes from a mongo collection
func MongoDropIndex(dbName, collectionName string, indexes []string) error {
	client := GetMongo()
	db := client.Database(dbName)               // set database name
	collection := db.Collection(collectionName) // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// drop index
	return collection.DropIndex(ctx, indexes)
}

// MongoDropAllIndexes - drop all indexes from a mongo collection except the index on the _id field
func MongoDropAllIndexes(dbName, collectionName string) error {
	client := GetMongo()
	db := client.Database(dbName)               // set database name
	collection := db.Collection(collectionName) // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// drop all indexes
	return collection.DropAllIndexes(ctx)
}
