package model

import (
	"reflect"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Geocoding - struct for address
type Geocoding struct {
	ID               primitive.ObjectID `json:"id" bson:"_id"`
	FormattedAddress string             `json:"formattedAddress,omitempty" bson:"formattedAddress,omitempty"`
	StreetName       string             `json:"streetName,omitempty" bson:"streetName,omitempty"`
	HouseNumber      string             `json:"houseNumber,omitempty" bson:"houseNumber,omitempty"`
	PostalCode       string             `json:"postalCode,omitempty" bson:"postalCode,omitempty"`
	County           string             `json:"county,omitempty" bson:"county,omitempty"`
	City             string             `json:"city,omitempty" bson:"city,omitempty"`
	State            string             `json:"state,omitempty" bson:"state,omitempty"`
	StateCode        string             `json:"stateCode,omitempty" bson:"stateCode,omitempty"`
	Country          string             `json:"country,omitempty" bson:"country,omitempty"`
	CountryCode      string             `json:"countryCode,omitempty" bson:"countryCode,omitempty"`
	Geometry         Geometry           `json:"geometry,omitempty" bson:"inline"`
}

// Geometry - struct for latitude and longitude
type Geometry struct {
	Latitude  float64 `json:"lat,omitempty" bson:"lat,omitempty"`
	Longitude float64 `json:"lng,omitempty" bson:"lng,omitempty"`
}

// IsEmpty - check empty struct
func (s Geocoding) IsEmpty() bool {
	return reflect.DeepEqual(s, Geocoding{})
}
