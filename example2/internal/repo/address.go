package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/pilinux/gorest/example2/internal/database/model"
)

// AddressRepo provides methods for address-related MongoDB operations.
type AddressRepo struct {
	db *mongo.Client
}

// NewAddressRepo returns a new AddressRepo.
func NewAddressRepo(conn *mongo.Client) *AddressRepo {
	return &AddressRepo{
		db: conn,
	}
}

// AddressRepository defines the contract for address data operations.
type AddressRepository interface {
	AddAddress(ctx context.Context, address *model.Geocoding) (*mongo.InsertOneResult, error)
	GetAddresses(ctx context.Context) ([]model.Geocoding, error)
	GetAddress(ctx context.Context, id bson.ObjectID) (*model.Geocoding, error)
	GetAddressByFilter(ctx context.Context, address *model.Geocoding, addDocIDInFilter bool) (*model.Geocoding, error)
	UpdateAddressFields(ctx context.Context, address *model.Geocoding) error
	DeleteAddress(ctx context.Context, id bson.ObjectID) error
}

// Compile-time check:
var _ AddressRepository = (*AddressRepo)(nil)

// dbName returns the database name for addresses.
func (r *AddressRepo) dbName() string {
	return "map"
}

// collName returns the collection name for addresses.
func (r *AddressRepo) collName() string {
	return "geocodes"
}

// coll returns the MongoDB collection for addresses.
func (r *AddressRepo) coll() *mongo.Collection {
	return r.db.Database(r.dbName()).Collection(r.collName())
}

// AddAddress inserts a new address into the MongoDB "geocodes" collection.
func (r *AddressRepo) AddAddress(ctx context.Context, address *model.Geocoding) (*mongo.InsertOneResult, error) {
	address.ID = bson.NewObjectID()
	return r.coll().InsertOne(ctx, address)
}

// GetAddresses retrieves all addresses from the MongoDB "geocodes" collection.
func (r *AddressRepo) GetAddresses(ctx context.Context) ([]model.Geocoding, error) {
	var addresses []model.Geocoding

	cursor, err := r.coll().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	if err := cursor.All(ctx, &addresses); err != nil {
		return nil, err
	}
	return addresses, nil
}

// GetAddress retrieves a specific address by its ID from the MongoDB "geocodes" collection.
func (r *AddressRepo) GetAddress(ctx context.Context, id bson.ObjectID) (*model.Geocoding, error) {
	var address model.Geocoding

	err := r.coll().FindOne(ctx, bson.M{"_id": id}).Decode(&address)
	if err != nil {
		return nil, err
	}
	return &address, nil
}

// GetAddressByFilter retrieves an address based on a filter from the MongoDB "geocodes" collection.
func (r *AddressRepo) GetAddressByFilter(ctx context.Context, address *model.Geocoding, addDocIDInFilter bool) (*model.Geocoding, error) {
	filter := addressFilter(address, addDocIDInFilter)

	err := r.coll().FindOne(ctx, filter).Decode(address)
	if err != nil {
		return nil, err
	}
	return address, nil
}

// UpdateAddress updates an existing address in the MongoDB "geocodes" collection.
func (r *AddressRepo) UpdateAddress(ctx context.Context, address *model.Geocoding) error {
	if address == nil || address.ID.IsZero() {
		return mongo.ErrNoDocuments
	}

	filter := bson.M{"_id": address.ID}
	update := bson.M{"$set": address}

	res, err := r.coll().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// DeleteFieldsFromAddress deletes specific fields from an address in the MongoDB "geocodes" collection.
func (r *AddressRepo) DeleteFieldsFromAddress(ctx context.Context, address *model.Geocoding, fields ...string) error {
	if address == nil || address.ID.IsZero() {
		return mongo.ErrNoDocuments
	}

	filter := bson.M{"_id": address.ID}
	unset := bson.M{}
	for _, field := range fields {
		unset[field] = ""
	}
	update := bson.M{"$unset": unset}

	res, err := r.coll().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// UpdateAddressFields updates an existing address (adding or deleting fields as necessary) in the MongoDB "geocodes" collection.
func (r *AddressRepo) UpdateAddressFields(ctx context.Context, address *model.Geocoding) error {
	if address == nil || address.ID.IsZero() {
		return mongo.ErrNoDocuments
	}

	filter := bson.M{"_id": address.ID}
	setFields := bson.M{}
	unsetFields := bson.M{}

	// helper to set/unset string fields
	setOrUnset := func(fieldName, value string) {
		if value != "" {
			setFields[fieldName] = value
		} else {
			unsetFields[fieldName] = ""
		}
	}

	setOrUnset("formattedAddress", address.FormattedAddress)
	setOrUnset("streetName", address.StreetName)
	setOrUnset("houseNumber", address.HouseNumber)
	setOrUnset("postalCode", address.PostalCode)
	setOrUnset("county", address.County)
	setOrUnset("city", address.City)
	setOrUnset("state", address.State)
	setOrUnset("stateCode", address.StateCode)
	setOrUnset("country", address.Country)
	setOrUnset("countryCode", address.CountryCode)

	// geometry fields
	if address.Geometry != nil {
		if address.Geometry.Latitude != nil {
			setFields["latitude"] = *address.Geometry.Latitude
		} else {
			unsetFields["latitude"] = ""
		}
		if address.Geometry.Longitude != nil {
			setFields["longitude"] = *address.Geometry.Longitude
		} else {
			unsetFields["longitude"] = ""
		}
	} else {
		unsetFields["latitude"] = ""
		unsetFields["longitude"] = ""
	}

	update := bson.M{}
	if len(setFields) > 0 {
		update["$set"] = setFields
	}
	if len(unsetFields) > 0 {
		update["$unset"] = unsetFields
	}

	res, err := r.coll().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// DeleteAddress deletes an address by its ID from the MongoDB "geocodes" collection.
func (r *AddressRepo) DeleteAddress(ctx context.Context, id bson.ObjectID) error {
	if id.IsZero() {
		return mongo.ErrNoDocuments
	}

	filter := bson.M{"_id": id}
	res, err := r.coll().DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// addressFilter builds a MongoDB search filter for the given address fields.
func addressFilter(address *model.Geocoding, addDocIDInFilter bool) bson.M {
	filter := bson.M{}
	if address == nil {
		return filter
	}

	if addDocIDInFilter && !address.ID.IsZero() {
		filter["_id"] = bson.M{"$eq": address.ID}
	}
	if address.FormattedAddress != "" {
		filter["formattedAddress"] = bson.M{"$eq": address.FormattedAddress}
	}
	if address.StreetName != "" {
		filter["streetName"] = bson.M{"$eq": address.StreetName}
	}
	if address.HouseNumber != "" {
		filter["houseNumber"] = bson.M{"$eq": address.HouseNumber}
	}
	if address.PostalCode != "" {
		filter["postalCode"] = bson.M{"$eq": address.PostalCode}
	}
	if address.County != "" {
		filter["county"] = bson.M{"$eq": address.County}
	}
	if address.City != "" {
		filter["city"] = bson.M{"$eq": address.City}
	}
	if address.State != "" {
		filter["state"] = bson.M{"$eq": address.State}
	}
	if address.StateCode != "" {
		filter["stateCode"] = bson.M{"$eq": address.StateCode}
	}
	if address.Country != "" {
		filter["country"] = bson.M{"$eq": address.Country}
	}
	if address.CountryCode != "" {
		filter["countryCode"] = bson.M{"$eq": address.CountryCode}
	}
	if address.Geometry != nil {
		if address.Geometry.Latitude != nil {
			filter["latitude"] = bson.M{"$eq": *address.Geometry.Latitude}
		}
		if address.Geometry.Longitude != nil {
			filter["longitude"] = bson.M{"$eq": *address.Geometry.Longitude}
		}
	}

	return filter
}
