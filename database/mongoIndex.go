package database

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// MongoCreateIndex creates one index for a mongo collection.
func MongoCreateIndex(dbName, collectionName string, index mongo.IndexModel) error {
	client := GetMongo()
	if client == nil {
		return errors.New("mongo client is not initialized")
	}

	db := client.Database(dbName)               // set database name
	collection := db.Collection(collectionName) // set collection name

	// guard: the driver requires an order-preserving key spec (e.g. bson.D)
	if index.Keys == nil {
		return errors.New("index keys must not be nil")
	}

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// create index
	_, err := collection.Indexes().CreateOne(ctx, index)
	return err
}

// MongoCreateIndexes creates multiple indexes for a mongo collection.
func MongoCreateIndexes(dbName, collectionName string, indexes []mongo.IndexModel) error {
	client := GetMongo()
	if client == nil {
		return errors.New("mongo client is not initialized")
	}

	db := client.Database(dbName)               // set database name
	collection := db.Collection(collectionName) // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// create indexes
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// MongoDropIndex drops one or multiple indexes from a mongo collection.
func MongoDropIndex(dbName, collectionName string, indexes []string) error {
	client := GetMongo()
	if client == nil {
		return errors.New("mongo client is not initialized")
	}

	db := client.Database(dbName)               // set database name
	collection := db.Collection(collectionName) // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// drop index
	for _, name := range indexes {
		if name == "" {
			continue
		}
		if name == "_id_" {
			// _id_ is mandatory and cannot be dropped.
			continue
		}

		// otherwise, try to drop by name first (supports full names, custom names like "myIndex").
		if err := collection.Indexes().DropOne(ctx, name); err == nil {
			continue
		}

		// fallback: treat it as a field spec (e.g. "countryCode" or "-countryCode") and drop by key.
		field := name
		order := int32(1)
		if strings.HasPrefix(name, "-") {
			order = -1
			field = strings.TrimPrefix(name, "-")
		}
		if field == "" {
			continue
		}
		if err := collection.Indexes().DropWithKey(ctx, bson.D{{Key: field, Value: order}}); err != nil {
			return err
		}
	}
	return nil
}

// MongoDropAllIndexes drops all indexes from a mongo collection except the index on the _id field.
func MongoDropAllIndexes(dbName, collectionName string) error {
	client := GetMongo()
	if client == nil {
		return errors.New("mongo client is not initialized")
	}

	db := client.Database(dbName)               // set database name
	collection := db.Collection(collectionName) // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// drop all indexes
	return collection.Indexes().DropAll(ctx)
}
