package model

type Theme struct {
	ID              uint    `json:"id"`
	PrimaryColor    string  `json:"primaryColor"`
	SecondaryColor  string  `json:"secondaryColor"`
	BackgroundColor string  `json:"backgroundColor"`
	BackgroundImage *string `json:"backgroundImage"`
}
