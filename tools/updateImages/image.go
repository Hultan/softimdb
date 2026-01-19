package main

// image represents a movie image.
type Poster struct {
	Id   int    `gorm:"column:id;primary_key"`
	Data []byte `gorm:"column:image;"`
}

// TableName returns the name of the table.
func (i *Poster) TableName() string {
	return "image"
}
