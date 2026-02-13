package model

import (
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Geocoding represents an address with geocoding information.
type Geocoding struct {
	ID               bson.ObjectID `json:"id" bson:"_id,omitempty"`
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

// Trim trims leading and trailing spaces from the address fields.
func (s *Geocoding) Trim() {
	if s.FormattedAddress != nil {
		*s.FormattedAddress = strings.TrimSpace(*s.FormattedAddress)
	}
	if s.StreetName != nil {
		*s.StreetName = strings.TrimSpace(*s.StreetName)
	}
	if s.HouseNumber != nil {
		*s.HouseNumber = strings.TrimSpace(*s.HouseNumber)
	}
	if s.PostalCode != nil {
		*s.PostalCode = strings.TrimSpace(*s.PostalCode)
	}
	if s.County != nil {
		*s.County = strings.TrimSpace(*s.County)
	}
	if s.City != nil {
		*s.City = strings.TrimSpace(*s.City)
	}
	if s.State != nil {
		*s.State = strings.TrimSpace(*s.State)
	}
	if s.StateCode != nil {
		*s.StateCode = strings.TrimSpace(*s.StateCode)
	}
	if s.Country != nil {
		*s.Country = strings.TrimSpace(*s.Country)
	}
	if s.CountryCode != nil {
		*s.CountryCode = strings.TrimSpace(*s.CountryCode)
	}
}

// Geometry represents latitude and longitude coordinates.
type Geometry struct {
	Latitude  *float64 `json:"latitude,omitempty" bson:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty" bson:"longitude,omitempty"`
}

// IsEmpty checks whether the Geocoding struct is empty.
func (s *Geocoding) IsEmpty() bool {
	return s.FormattedAddress == nil &&
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
