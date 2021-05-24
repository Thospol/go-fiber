package models

// UserSession user session
type UserSession struct {
	Id          uint   `json:"userId"`
	AccessUUID  string `json:"accessUUID"`
	RefreshUUID string `json:"refreshUUID"`
}
