package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	gdatabase "github.com/pilinux/gorest/database"
	gmodel "github.com/pilinux/gorest/database/model"

	"github.com/pilinux/gorest/example/database/model"
)

// MongoCreateOne inserts a single geocoding document into MongoDB.
func MongoCreateOne(data model.Geocoding) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// remove all leading and trailing white spaces
	data = mongoTrimSpace(data)
	if data.IsEmpty() {
		httpResponse.Message = "empty body"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// generate a new ObjectID
	data.ID = bson.NewObjectID()

	client := gdatabase.GetMongo()
	db := client.Database("map")            // set database name
	collection := db.Collection("geocodes") // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// insert one document
	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		log.WithError(err).Error("error code: 1401")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = data
	httpStatusCode = http.StatusCreated
	return
}

// MongoGetAll retrieves all geocoding documents from MongoDB.
func MongoGetAll() (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	client := gdatabase.GetMongo()
	db := client.Database("map")            // set database name
	collection := db.Collection("geocodes") // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data := []model.Geocoding{}
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.WithError(err).Error("error code: 1411")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()
	if err := cursor.All(ctx, &data); err != nil {
		log.WithError(err).Error("error code: 1412")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if len(data) == 0 {
		httpResponse.Message = "no record found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = data
	httpStatusCode = http.StatusOK
	return
}

// MongoGetByID retrieves a geocoding document by ObjectID.
func MongoGetByID(id string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		httpResponse.Message = "invalid id"
		httpStatusCode = http.StatusBadRequest
		return
	}

	client := gdatabase.GetMongo()
	db := client.Database("map")            // set database name
	collection := db.Collection("geocodes") // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data := model.Geocoding{}
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&data)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			httpResponse.Message = "document not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("error code: 1421")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = data
	httpStatusCode = http.StatusOK
	return
}

// MongoGetByFilter retrieves geocoding documents matching the provided fields.
func MongoGetByFilter(req model.Geocoding) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// remove all leading and trailing white spaces
	req = mongoTrimSpace(req)

	// search filter
	filter := mongoFilter(req, true)

	if len(filter) == 0 {
		httpResponse.Message = "received empty json payload"
		httpStatusCode = http.StatusBadRequest
		return
	}

	client := gdatabase.GetMongo()
	db := client.Database("map")            // set database name
	collection := db.Collection("geocodes") // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data := []model.Geocoding{}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.WithError(err).Error("error code: 1431")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()
	if err := cursor.All(ctx, &data); err != nil {
		log.WithError(err).Error("error code: 1432")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if len(data) == 0 {
		httpResponse.Message = "no record found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = data
	httpStatusCode = http.StatusOK
	return
}

// MongoUpdateByID updates a geocoding document by ObjectID.
func MongoUpdateByID(req model.Geocoding) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	if req.ID.IsZero() {
		httpResponse.Message = "document ID is missing"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// remove all leading and trailing white spaces
	req = mongoTrimSpace(req)

	// search filter
	filter := bson.M{
		"_id": bson.M{"$eq": req.ID},
	}

	client := gdatabase.GetMongo()
	db := client.Database("map")            // set database name
	collection := db.Collection("geocodes") // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// create the update
	// https://docs.mongodb.com/manual/reference/operator/update/
	update := bson.M{
		"$set": req,
	}

	// find one result and update it
	res, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			httpResponse.Message = "document not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("error code: 1441")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if res.MatchedCount == 0 {
		httpResponse.Message = "document not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = req
	httpStatusCode = http.StatusOK
	return
}

// MongoDeleteFieldByID unsets specific fields on a geocoding document by ObjectID.
func MongoDeleteFieldByID(req model.Geocoding) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	if req.ID.IsZero() {
		httpResponse.Message = "document ID is missing"
		httpStatusCode = http.StatusBadRequest
		return
	}

	deleteFields := mongoFilter(req, false)

	// search filter
	filter := bson.M{
		"_id": bson.M{"$eq": req.ID},
	}

	client := gdatabase.GetMongo()
	db := client.Database("map")            // set database name
	collection := db.Collection("geocodes") // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// create the update
	// https://docs.mongodb.com/manual/reference/operator/update/
	update := bson.M{
		"$unset": deleteFields,
	}

	// find one result and update it
	res, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			httpResponse.Message = "document not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("error code: 1451")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if res.MatchedCount == 0 {
		httpResponse.Message = "document not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = req
	httpStatusCode = http.StatusOK
	return
}

// MongoDeleteByID deletes a geocoding document by ObjectID.
func MongoDeleteByID(id string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		httpResponse.Message = "invalid id"
		httpStatusCode = http.StatusBadRequest
		return
	}

	client := gdatabase.GetMongo()
	db := client.Database("map")            // set database name
	collection := db.Collection("geocodes") // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			httpResponse.Message = "document not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("error code: 1461")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	if res.DeletedCount == 0 {
		httpResponse.Message = "document not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = "document deleted successfully"
	httpStatusCode = http.StatusOK
	return
}

// mongoTrimSpace removes all leading and trailing white spaces from geocoding fields.
func mongoTrimSpace(geocoding model.Geocoding) model.Geocoding {
	geocoding.FormattedAddress = strings.TrimSpace(geocoding.FormattedAddress)
	geocoding.StreetName = strings.TrimSpace(geocoding.StreetName)
	geocoding.HouseNumber = strings.TrimSpace(geocoding.HouseNumber)
	geocoding.PostalCode = strings.TrimSpace(geocoding.PostalCode)
	geocoding.County = strings.TrimSpace(geocoding.County)
	geocoding.City = strings.TrimSpace(geocoding.City)
	geocoding.State = strings.TrimSpace(geocoding.State)
	geocoding.StateCode = strings.TrimSpace(geocoding.StateCode)
	geocoding.Country = strings.TrimSpace(geocoding.Country)
	geocoding.CountryCode = strings.TrimSpace(geocoding.CountryCode)

	return geocoding
}

// mongoFilter builds a search filter for MongoDB queries.
func mongoFilter(geocoding model.Geocoding, addDocIDInFilter bool) bson.M {
	filter := bson.M{}

	if addDocIDInFilter {
		if !geocoding.ID.IsZero() {
			filter["_id"] = bson.M{"$eq": geocoding.ID}
		}
	}
	if geocoding.FormattedAddress != "" {
		filter["formattedAddress"] = bson.M{"$eq": geocoding.FormattedAddress}
	}
	if geocoding.StreetName != "" {
		filter["streetName"] = bson.M{"$eq": geocoding.StreetName}
	}
	if geocoding.HouseNumber != "" {
		filter["houseNumber"] = bson.M{"$eq": geocoding.HouseNumber}
	}
	if geocoding.PostalCode != "" {
		filter["postalCode"] = bson.M{"$eq": geocoding.PostalCode}
	}
	if geocoding.County != "" {
		filter["county"] = bson.M{"$eq": geocoding.County}
	}
	if geocoding.City != "" {
		filter["city"] = bson.M{"$eq": geocoding.City}
	}
	if geocoding.State != "" {
		filter["state"] = bson.M{"$eq": geocoding.State}
	}
	if geocoding.StateCode != "" {
		filter["stateCode"] = bson.M{"$eq": geocoding.StateCode}
	}
	if geocoding.Country != "" {
		filter["country"] = bson.M{"$eq": geocoding.Country}
	}
	if geocoding.CountryCode != "" {
		filter["countryCode"] = bson.M{"$eq": geocoding.CountryCode}
	}
	if geocoding.Geometry.Latitude != 0 {
		filter["lat"] = bson.M{"$eq": geocoding.Geometry.Latitude}
	}
	if geocoding.Geometry.Longitude != 0 {
		filter["lng"] = bson.M{"$eq": geocoding.Geometry.Longitude}
	}

	return filter
}
