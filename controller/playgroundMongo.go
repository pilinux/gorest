package controller

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/lib/renderer"
)

// Geocoding - struct for address
type Geocoding struct {
	ID               primitive.ObjectID `json:"id" bson:"_id"`
	FormattedAddress string             `json:"formatted_address,omitempty" bson:"formatted_address,omitempty"`
	StreetName       string             `json:"street_name,omitempty" bson:"street_name,omitempty"`
	HouseNumber      string             `json:"house_number,omitempty" bson:"house_number,omitempty"`
	PostalCode       string             `json:"postal_code,omitempty" bson:"postal_code,omitempty"`
	County           string             `json:"county,omitempty" bson:"county,omitempty"`
	City             string             `json:"city,omitempty" bson:"city,omitempty"`
	State            string             `json:"state,omitempty" bson:"state,omitempty"`
	StateCode        string             `json:"state_code,omitempty" bson:"state_code,omitempty"`
	Country          string             `json:"country,omitempty" bson:"country,omitempty"`
	CountryCode      string             `json:"country_code,omitempty" bson:"country_code,omitempty"`
	Geometry         Geometry           `json:"geometry,omitempty" bson:"inline"`
}

// Geometry - struct for latitude and longitude
type Geometry struct {
	Latitude  float64 `json:"lat,omitempty" bson:"lat,omitempty"`
	Longitude float64 `json:"lng,omitempty" bson:"lng,omitempty"`
}

// MongoCreateOne - create one document
func MongoCreateOne(c *gin.Context) {
	data := Geocoding{}
	if err := c.ShouldBindJSON(&data); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	// remove all leading and trailing white spaces
	data = MongoTrimSpace(data)
	if data.isEmpty() {
		renderer.Render(c, gin.H{"msg": "empty body"}, http.StatusBadRequest)
		return
	}

	// generate a new ObjectID
	data.ID = primitive.NewObjectID()

	client := database.GetMongo()
	db := client.Database("map")            // set database name
	collection := db.Collection("geocodes") // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// insert one document
	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		log.WithError(err).Error("error code: 1401")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	renderer.Render(c, data, http.StatusCreated)
}

// MongoGetAll - get all documents
func MongoGetAll(c *gin.Context) {
	client := database.GetMongo()
	db := client.Database("map")            // set database name
	collection := db.Collection("geocodes") // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data := []Geocoding{}
	err := collection.Find(ctx, bson.M{}).All(&data)
	if err != nil {
		log.WithError(err).Error("error code: 1411")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	if len(data) == 0 {
		renderer.Render(c, gin.H{"msg": "no record found"}, http.StatusNotFound)
		return
	}

	renderer.Render(c, data, http.StatusOK)
}

// MongoGetByID - find one document by ID
func MongoGetByID(c *gin.Context) {
	id := strings.TrimSpace(c.Params.ByName("id"))
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		renderer.Render(c, gin.H{"msg": "invalid id"}, http.StatusBadRequest)
		return
	}

	client := database.GetMongo()
	db := client.Database("map")            // set database name
	collection := db.Collection("geocodes") // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data := Geocoding{}
	err = collection.Find(ctx, bson.M{"_id": objID}).One(&data)
	if err != nil {
		renderer.Render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}

	renderer.Render(c, data, http.StatusOK)
}

// MongoGetByFilter - find documents using filter
func MongoGetByFilter(c *gin.Context) {
	req := Geocoding{}
	if err := c.ShouldBindJSON(&req); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	// remove all leading and trailing white spaces
	req = MongoTrimSpace(req)

	// search filter
	filter := MongoFilter(req, true)

	if len(filter) == 0 {
		renderer.Render(c, gin.H{"msg": "received empty json payload"}, http.StatusBadRequest)
		return
	}

	client := database.GetMongo()
	db := client.Database("map")            // set database name
	collection := db.Collection("geocodes") // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data := []Geocoding{}

	err := collection.Find(ctx, filter).All(&data)
	if err != nil {
		log.WithError(err).Error("error code: 1421")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	if len(data) == 0 {
		renderer.Render(c, gin.H{"msg": "no record found"}, http.StatusNotFound)
		return
	}

	renderer.Render(c, data, http.StatusOK)
}

// MongoUpdateByID - update a document
// edit existing fields
// add new fields
// do not remove any existing field
func MongoUpdateByID(c *gin.Context) {
	req := Geocoding{}
	if err := c.ShouldBindJSON(&req); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}
	if req.ID.IsZero() {
		renderer.Render(c, gin.H{"msg": "document ID is missing"}, http.StatusBadRequest)
		return
	}

	// remove all leading and trailing white spaces
	req = MongoTrimSpace(req)

	// search filter
	filter := bson.M{
		"_id": bson.M{"$eq": req.ID},
	}

	client := database.GetMongo()
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
		renderer.Render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}

	renderer.Render(c, req, http.StatusOK)
}

// MongoDeleteFieldByID - delete existing field(s) from a document
func MongoDeleteFieldByID(c *gin.Context) {
	req := Geocoding{}
	if err := c.ShouldBindJSON(&req); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}
	if req.ID.IsZero() {
		renderer.Render(c, gin.H{"msg": "document ID is missing"}, http.StatusBadRequest)
		return
	}

	deleteFields := MongoFilter(req, false)

	// search filter
	filter := bson.M{
		"_id": bson.M{"$eq": req.ID},
	}

	client := database.GetMongo()
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
		renderer.Render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}

	renderer.Render(c, gin.H{"msg": "fields removed from the document"}, http.StatusOK)
}

// MongoDeleteByID - delete one document by ID
func MongoDeleteByID(c *gin.Context) {
	id := strings.TrimSpace(c.Params.ByName("id"))
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		renderer.Render(c, gin.H{"msg": "invalid id"}, http.StatusBadRequest)
		return
	}

	client := database.GetMongo()
	db := client.Database("map")            // set database name
	collection := db.Collection("geocodes") // set collection name

	// set max TTL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = collection.Remove(ctx, bson.M{"_id": objID})
	if err != nil {
		renderer.Render(c, gin.H{"msg": "document not found/cannot be deleted"}, http.StatusNotFound)
		return
	}

	renderer.Render(c, gin.H{"msg": "document deleted successfully"}, http.StatusOK)
}

// MongoTrimSpace - remove all leading and trailing white spaces
func MongoTrimSpace(geocoding Geocoding) Geocoding {
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

// MongoFilter - search filter
func MongoFilter(geocoding Geocoding, addDocIDInFilter bool) bson.M {
	filter := bson.M{}

	if addDocIDInFilter {
		if !geocoding.ID.IsZero() {
			filter["_id"] = bson.M{"$eq": geocoding.ID}
		}
	}
	if geocoding.FormattedAddress != "" {
		filter["formatted_address"] = bson.M{"$eq": geocoding.FormattedAddress}
	}
	if geocoding.StreetName != "" {
		filter["street_name"] = bson.M{"$eq": geocoding.StreetName}
	}
	if geocoding.HouseNumber != "" {
		filter["house_number"] = bson.M{"$eq": geocoding.HouseNumber}
	}
	if geocoding.PostalCode != "" {
		filter["postal_code"] = bson.M{"$eq": geocoding.PostalCode}
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
		filter["state_code"] = bson.M{"$eq": geocoding.StateCode}
	}
	if geocoding.Country != "" {
		filter["country"] = bson.M{"$eq": geocoding.Country}
	}
	if geocoding.CountryCode != "" {
		filter["country_code"] = bson.M{"$eq": geocoding.CountryCode}
	}
	if geocoding.Geometry.Latitude != 0 {
		filter["lat"] = bson.M{"$eq": geocoding.Geometry.Latitude}
	}
	if geocoding.Geometry.Longitude != 0 {
		filter["lng"] = bson.M{"$eq": geocoding.Geometry.Longitude}
	}

	return filter
}

// isEmpty - check empty struct
func (s Geocoding) isEmpty() bool {
	return reflect.DeepEqual(s, Geocoding{})
}
