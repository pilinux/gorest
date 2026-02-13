package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

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

	if err := mongoValidateGeocodingQuery(req); err != nil {
		httpResponse.Message = "invalid query payload"
		httpStatusCode = http.StatusBadRequest
		return
	}

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

	if err := mongoValidateGeocodingQuery(req); err != nil {
		httpResponse.Message = "invalid query payload"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// search filter
	filter := bson.M{
		"_id": req.ID,
	}

	client := gdatabase.GetMongo()
	db := client.Database("map")            // set database name
	collection := db.Collection("geocodes") // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// create the update
	// https://docs.mongodb.com/manual/reference/operator/update/
	setFields := mongoSetFields(req)
	if len(setFields) == 0 {
		httpResponse.Message = "received empty json payload"
		httpStatusCode = http.StatusBadRequest
		return
	}
	update := bson.M{
		"$set": setFields,
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

	// fetch the updated document
	var updatedDoc model.Geocoding
	err = collection.FindOne(ctx, filter).Decode(&updatedDoc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			httpResponse.Message = "document not found after update"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("error code: 1442")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = updatedDoc
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

	deleteFields := mongoUnsetFields(req)
	if len(deleteFields) == 0 {
		httpResponse.Message = "received empty json payload"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// search filter
	filter := bson.M{
		"_id": req.ID,
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
	if geocoding.FormattedAddress != nil {
		*geocoding.FormattedAddress = strings.TrimSpace(*geocoding.FormattedAddress)
	}
	if geocoding.StreetName != nil {
		*geocoding.StreetName = strings.TrimSpace(*geocoding.StreetName)
	}
	if geocoding.HouseNumber != nil {
		*geocoding.HouseNumber = strings.TrimSpace(*geocoding.HouseNumber)
	}
	if geocoding.PostalCode != nil {
		*geocoding.PostalCode = strings.TrimSpace(*geocoding.PostalCode)
	}
	if geocoding.County != nil {
		*geocoding.County = strings.TrimSpace(*geocoding.County)
	}
	if geocoding.City != nil {
		*geocoding.City = strings.TrimSpace(*geocoding.City)
	}
	if geocoding.State != nil {
		*geocoding.State = strings.TrimSpace(*geocoding.State)
	}
	if geocoding.StateCode != nil {
		*geocoding.StateCode = strings.TrimSpace(*geocoding.StateCode)
	}
	if geocoding.Country != nil {
		*geocoding.Country = strings.TrimSpace(*geocoding.Country)
	}
	if geocoding.CountryCode != nil {
		*geocoding.CountryCode = strings.TrimSpace(*geocoding.CountryCode)
	}

	return geocoding
}

// mongoFilter builds a search filter for MongoDB queries.
func mongoFilter(geocoding model.Geocoding, addDocIDInFilter bool) bson.M {
	filter := bson.M{}

	if addDocIDInFilter {
		if !geocoding.ID.IsZero() {
			filter["_id"] = geocoding.ID
		}
	}
	if geocoding.FormattedAddress != nil && *geocoding.FormattedAddress != "" {
		filter["formattedAddress"] = *geocoding.FormattedAddress
	}
	if geocoding.StreetName != nil && *geocoding.StreetName != "" {
		filter["streetName"] = *geocoding.StreetName
	}
	if geocoding.HouseNumber != nil && *geocoding.HouseNumber != "" {
		filter["houseNumber"] = *geocoding.HouseNumber
	}
	if geocoding.PostalCode != nil && *geocoding.PostalCode != "" {
		filter["postalCode"] = *geocoding.PostalCode
	}
	if geocoding.County != nil && *geocoding.County != "" {
		filter["county"] = *geocoding.County
	}
	if geocoding.City != nil && *geocoding.City != "" {
		filter["city"] = *geocoding.City
	}
	if geocoding.State != nil && *geocoding.State != "" {
		filter["state"] = *geocoding.State
	}
	if geocoding.StateCode != nil && *geocoding.StateCode != "" {
		filter["stateCode"] = *geocoding.StateCode
	}
	if geocoding.Country != nil && *geocoding.Country != "" {
		filter["country"] = *geocoding.Country
	}
	if geocoding.CountryCode != nil && *geocoding.CountryCode != "" {
		filter["countryCode"] = *geocoding.CountryCode
	}
	if geocoding.Geometry != nil && geocoding.Geometry.Latitude != nil {
		filter["lat"] = *geocoding.Geometry.Latitude
	}
	if geocoding.Geometry != nil && geocoding.Geometry.Longitude != nil {
		filter["lng"] = *geocoding.Geometry.Longitude
	}

	return filter
}

func mongoValidateGeocodingQuery(geocoding model.Geocoding) error {
	const maxLen = 256

	fields := []*string{
		geocoding.FormattedAddress,
		geocoding.StreetName,
		geocoding.HouseNumber,
		geocoding.PostalCode,
		geocoding.County,
		geocoding.City,
		geocoding.State,
		geocoding.StateCode,
		geocoding.Country,
		geocoding.CountryCode,
	}
	for _, v := range fields {
		if v == nil || *v == "" {
			continue
		}
		if len(*v) > maxLen {
			return errors.New("field too long")
		}
		if strings.ContainsRune(*v, '\x00') {
			return errors.New("field contains null byte")
		}
		if !utf8.ValidString(*v) {
			return errors.New("field contains invalid utf-8")
		}
	}

	if geocoding.Geometry != nil && geocoding.Geometry.Latitude != nil && *geocoding.Geometry.Latitude != 0 && (*geocoding.Geometry.Latitude < -90 || *geocoding.Geometry.Latitude > 90) {
		return errors.New("latitude out of range")
	}
	if geocoding.Geometry != nil && geocoding.Geometry.Longitude != nil && *geocoding.Geometry.Longitude != 0 && (*geocoding.Geometry.Longitude < -180 || *geocoding.Geometry.Longitude > 180) {
		return errors.New("longitude out of range")
	}

	return nil
}

func mongoSetFields(geocoding model.Geocoding) bson.M {
	setFields := bson.M{}

	if geocoding.FormattedAddress != nil {
		setFields["formattedAddress"] = *geocoding.FormattedAddress
	}
	if geocoding.StreetName != nil {
		setFields["streetName"] = *geocoding.StreetName
	}
	if geocoding.HouseNumber != nil {
		setFields["houseNumber"] = *geocoding.HouseNumber
	}
	if geocoding.PostalCode != nil {
		setFields["postalCode"] = *geocoding.PostalCode
	}
	if geocoding.County != nil {
		setFields["county"] = *geocoding.County
	}
	if geocoding.City != nil {
		setFields["city"] = *geocoding.City
	}
	if geocoding.State != nil {
		setFields["state"] = *geocoding.State
	}
	if geocoding.StateCode != nil {
		setFields["stateCode"] = *geocoding.StateCode
	}
	if geocoding.Country != nil {
		setFields["country"] = *geocoding.Country
	}
	if geocoding.CountryCode != nil {
		setFields["countryCode"] = *geocoding.CountryCode
	}
	if geocoding.Geometry != nil && geocoding.Geometry.Latitude != nil {
		setFields["lat"] = *geocoding.Geometry.Latitude
	}
	if geocoding.Geometry != nil && geocoding.Geometry.Longitude != nil {
		setFields["lng"] = *geocoding.Geometry.Longitude
	}

	return setFields
}

func mongoUnsetFields(geocoding model.Geocoding) bson.M {
	unsetFields := bson.M{}

	if geocoding.FormattedAddress != nil {
		unsetFields["formattedAddress"] = 1
	}
	if geocoding.StreetName != nil {
		unsetFields["streetName"] = 1
	}
	if geocoding.HouseNumber != nil {
		unsetFields["houseNumber"] = 1
	}
	if geocoding.PostalCode != nil {
		unsetFields["postalCode"] = 1
	}
	if geocoding.County != nil {
		unsetFields["county"] = 1
	}
	if geocoding.City != nil {
		unsetFields["city"] = 1
	}
	if geocoding.State != nil {
		unsetFields["state"] = 1
	}
	if geocoding.StateCode != nil {
		unsetFields["stateCode"] = 1
	}
	if geocoding.Country != nil {
		unsetFields["country"] = 1
	}
	if geocoding.CountryCode != nil {
		unsetFields["countryCode"] = 1
	}
	if geocoding.Geometry != nil && geocoding.Geometry.Latitude != nil {
		unsetFields["lat"] = 1
	}
	if geocoding.Geometry != nil && geocoding.Geometry.Longitude != nil {
		unsetFields["lng"] = 1
	}

	return unsetFields
}
