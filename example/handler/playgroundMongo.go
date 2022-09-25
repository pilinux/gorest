package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	gdatabase "github.com/pilinux/gorest/database"
	gmodel "github.com/pilinux/gorest/database/model"

	"github.com/pilinux/gorest/example/database/model"
)

// MongoCreateOne - handles jobs for controller.MongoCreateOne
func MongoCreateOne(data model.Geocoding) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// remove all leading and trailing white spaces
	data = mongoTrimSpace(data)
	if data.IsEmpty() {
		httpResponse.Message = "empty body"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// generate a new ObjectID
	data.ID = primitive.NewObjectID()

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

// MongoGetAll handles jobs for controller.MongoGetAll
func MongoGetAll() (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	client := gdatabase.GetMongo()
	db := client.Database("map")            // set database name
	collection := db.Collection("geocodes") // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data := []model.Geocoding{}
	err := collection.Find(ctx, bson.M{}).All(&data)
	if err != nil {
		log.WithError(err).Error("error code: 1411")
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

// MongoGetByID handles jobs for controller.MongoGetByID
func MongoGetByID(id string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	objID, err := primitive.ObjectIDFromHex(id)
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
	err = collection.Find(ctx, bson.M{"_id": objID}).One(&data)
	if err != nil {
		httpResponse.Message = "document not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = data
	httpStatusCode = http.StatusOK
	return
}

// MongoGetByFilter handles jobs for controller.MongoGetByFilter
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
	err := collection.Find(ctx, filter).All(&data)
	if err != nil {
		log.WithError(err).Error("error code: 1421")
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

// MongoUpdateByID handles jobs for controller.MongoUpdateByID
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
	err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		httpResponse.Message = "document not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = req
	httpStatusCode = http.StatusOK
	return
}

// MongoDeleteFieldByID handles jobs for controller.MongoDeleteFieldByID
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
	err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		httpResponse.Message = "document not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = req
	httpStatusCode = http.StatusOK
	return
}

// MongoDeleteByID handles jobs for controller.MongoDeleteByID
func MongoDeleteByID(id string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	objID, err := primitive.ObjectIDFromHex(id)
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

	err = collection.Remove(ctx, bson.M{"_id": objID})
	if err != nil {
		httpResponse.Message = "document not found/cannot be deleted"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = "document deleted successfully"
	httpStatusCode = http.StatusOK
	return
}

// mongoTrimSpace - remove all leading and trailing white spaces
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

// mongoFilter - search filter
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
