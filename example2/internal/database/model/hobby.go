package model

// Hobby model - `hobbies` table
type Hobby struct {
	HobbyID   uint64 `gorm:"primaryKey" json:"hobbyID,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty"`
	Hobby     string `json:"hobby,omitempty"`
	Users     []User `gorm:"many2many:user_hobbies" json:"-"`
}
