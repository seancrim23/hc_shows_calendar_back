package services

import "hc_shows_backend/models"

//the hc show calendar service will contain all functions that
//will interact with the database

// this is created as an interface so we can implement different db connections
// and their respective functions as needed
type HCShowCalendarService interface {
	//TODO add pagination (if needed) which should contain a member to hold the query filters
	GetShows(map[string]string) (*[]models.Show, error)
	GetShow(string) (*models.Show, error)
	CreateShow(models.Show, string) (*models.Show, error)
	UpdateShow(string, models.Show) (*models.Show, error)
	DeleteShow(string) error

	GetUser(string) (*models.User, error)
	GetUserShows(string) (*[]models.Show, error)
	CreateUser(models.User) (*models.User, error)
	UpdateUser(string, models.User) (*models.User, error)
	DeleteUser(string) error

	ResetPassword(string, string) error

	AuthUser(models.User) (string, error)
	CreateAuthObject(*models.Verification) error
	DeleteAuthObject(string) error
	ValidateAuthUser(string, string) error
}
