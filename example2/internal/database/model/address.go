package model

import (
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Geocoding - struct for address
type Geocoding struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
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
	Geometry         *Geometry          `json:"geometry,omitempty" bson:"inline"`
}

// Trim trims leading and trailing spaces from the address fields.
func (s *Geocoding) Trim() {
	s.FormattedAddress = strings.TrimSpace(s.FormattedAddress)
	s.StreetName = strings.TrimSpace(s.StreetName)
	s.HouseNumber = strings.TrimSpace(s.HouseNumber)
	s.PostalCode = strings.TrimSpace(s.PostalCode)
	s.County = strings.TrimSpace(s.County)
	s.City = strings.TrimSpace(s.City)
	s.State = strings.TrimSpace(s.State)
	s.StateCode = strings.TrimSpace(s.StateCode)
	s.Country = strings.TrimSpace(s.Country)
	s.CountryCode = strings.TrimSpace(s.CountryCode)
}

// Geometry - struct for latitude and longitude
type Geometry struct {
	Latitude  *float64 `json:"latitude,omitempty" bson:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty" bson:"longitude,omitempty"`
}

// IsEmpty - check empty struct
func (s *Geocoding) IsEmpty() bool {
	return s.FormattedAddress == "" &&
		s.StreetName == "" &&
		s.HouseNumber == "" &&
		s.PostalCode == "" &&
		s.County == "" &&
		s.City == "" &&
		s.State == "" &&
		s.StateCode == "" &&
		s.Country == "" &&
		s.CountryCode == "" &&
		(s.Geometry == nil || (s.Geometry.Latitude == nil && s.Geometry.Longitude == nil))
}
