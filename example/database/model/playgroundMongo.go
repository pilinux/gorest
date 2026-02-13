package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Geocoding represents an address for geocoding.
type Geocoding struct {
	ID               bson.ObjectID `json:"id" bson:"_id"`
	FormattedAddress *string       `json:"formattedAddress,omitempty" bson:"formattedAddress,omitempty"`
	StreetName       *string       `json:"streetName,omitempty" bson:"streetName,omitempty"`
	HouseNumber      *string       `json:"houseNumber,omitempty" bson:"houseNumber,omitempty"`
	PostalCode       *string       `json:"postalCode,omitempty" bson:"postalCode,omitempty"`
	County           *string       `json:"county,omitempty" bson:"county,omitempty"`
	City             *string       `json:"city,omitempty" bson:"city,omitempty"`
	State            *string       `json:"state,omitempty" bson:"state,omitempty"`
	StateCode        *string       `json:"stateCode,omitempty" bson:"stateCode,omitempty"`
	Country          *string       `json:"country,omitempty" bson:"country,omitempty"`
	CountryCode      *string       `json:"countryCode,omitempty" bson:"countryCode,omitempty"`
	Geometry         *Geometry     `json:"geometry,omitempty" bson:"inline"`
}

// Geometry represents latitude and longitude coordinates.
type Geometry struct {
	Latitude  *float64 `json:"lat,omitempty" bson:"lat,omitempty"`
	Longitude *float64 `json:"lng,omitempty" bson:"lng,omitempty"`
}

// IsEmpty checks whether the Geocoding struct is empty.
func (s Geocoding) IsEmpty() bool {
	return s.ID.IsZero() &&
		s.FormattedAddress == nil &&
		s.StreetName == nil &&
		s.HouseNumber == nil &&
		s.PostalCode == nil &&
		s.County == nil &&
		s.City == nil &&
		s.State == nil &&
		s.StateCode == nil &&
		s.Country == nil &&
		s.CountryCode == nil &&
		s.Geometry == nil
}
