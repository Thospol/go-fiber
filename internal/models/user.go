package models

// User user model
type User struct {
	Model
	Pronoun string `json:"pronoun"`
	Name    string `json:"name"`
}
