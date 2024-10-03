package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"hc_shows_backend/models"
	"hc_shows_backend/utils"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/fatih/structs"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/iterator"
)

// add variable to hold the db connection
// tbd any other variables that may be needed for the service
type FirestoreHCShowCalendarService struct {
	database *firestore.Client
	ctx      context.Context
}

func NewFirestoreHCShowCalendarService() (*FirestoreHCShowCalendarService, func(), error) {
	ctx := context.Background()

	if value := os.Getenv(utils.FIRESTORE_EMULATOR_HOST); value != "" {
		fmt.Println("using firestore emulator: " + value)
	}

	conf := &firebase.Config{ProjectID: os.Getenv(utils.GCP_PROJECT_ID)}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		fmt.Println("error making new firebase app: ", err)
		return nil, nil, err
	}
	database, err := app.Firestore(ctx)
	if err != nil {
		fmt.Println("error making firestore connection: ", err)
		return nil, nil, err
	}

	closeFunc := func() {
		database.Close()
	}

	return &FirestoreHCShowCalendarService{database: database, ctx: ctx}, closeFunc, nil
}

// TODO pagination at some point
func (f *FirestoreHCShowCalendarService) GetShows(showQueryFilters map[string]string) (*[]models.Show, error) {
	var shows []models.Show
	var q firestore.Query
	var iter *firestore.DocumentIterator
	//TODO clean up query building its really gross
	currentDate := time.Now()
	dateRangeDate := time.Now()
	if val, ok := showQueryFilters["date_range"]; ok {
		dateRangeDate = currentDate.AddDate(0, 0, utils.DateRangeMapping[val])
	}
	collection := f.database.Collection(utils.SHOW_COLLECTION)
	if len(showQueryFilters) == 0 {
		iter = collection.OrderBy("date", firestore.Asc).Documents(f.ctx)
	} else {
		index := 0
		for k, v := range showQueryFilters {
			if k != "date_range" {
				if index == 0 {
					q = collection.Where(k, "==", v)
				} else {
					q = q.Where(k, "==", v)
				}
				index++
			}
		}
		if _, ok := showQueryFilters["date_range"]; ok {
			q = q.Where("date", "<=", dateRangeDate).Where("date", ">=", currentDate)
		}
		iter = q.OrderBy("date", firestore.Asc).Documents(f.ctx)
	}
	for {
		s := models.Show{}
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println("error iterating over results")
			return nil, err
		}
		err = doc.DataTo(&s)
		if err != nil {
			fmt.Println("error getting all shows")
			fmt.Println(err)
			return nil, errors.New("error getting all shows")
		}
		shows = append(shows, s)
	}
	return &shows, nil
}

func (f *FirestoreHCShowCalendarService) GetShow(id string) (*models.Show, error) {
	dsnap, err := f.database.Collection(utils.SHOW_COLLECTION).Doc(id).Get(f.ctx)
	if err != nil {
		fmt.Println("error getting show")
		fmt.Println(err)
		return nil, errors.New("error getting show")
	}
	var s models.Show
	err = dsnap.DataTo(&s)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("error getting show")
	}
	return &s, nil
}

func (f *FirestoreHCShowCalendarService) CreateShow(show models.Show, username string) (*models.Show, error) {
	if show.Id == "" {
		newShowId := uuid.New()
		show.Id = newShowId.String()
	}
	show.Promoter = username
	_, err := f.database.Collection(utils.SHOW_COLLECTION).Doc(show.Id).Set(f.ctx, show)
	if err != nil {
		fmt.Println("error creating show")
		fmt.Println(err)
		return nil, errors.New("error creating show")
	}
	return &show, nil
}

func (f *FirestoreHCShowCalendarService) UpdateShow(id string, show models.Show) (*models.Show, error) {
	showFirestoreUpdateData := buildShowFirestoreUpdateData(show)
	_, err := f.database.Collection(utils.SHOW_COLLECTION).Doc(id).Update(f.ctx, showFirestoreUpdateData)
	if err != nil {
		fmt.Println("error updating show")
		fmt.Println(err)
		return nil, errors.New("error updating show")
	}

	return &show, nil
}

// TODO this whole function kinda sucks but works so make it better sometime
// is there a better way to get data from a request -> validate -> pass to update?
func buildShowFirestoreUpdateData(show models.Show) []firestore.Update {
	fireStoreUpdates := []firestore.Update{}
	//i dont like this but i dont think i have a better choice to
	//turn my go struct into something firestore will like
	showTempGenericMap := structs.Map(show)
	for i, v := range showTempGenericMap {
		if strings.ToLower(i) != "id" && strings.ToLower(i) != "promoter" {
			fireStoreUpdates = append(fireStoreUpdates, firestore.Update{Path: strings.ToLower(i), Value: v})
		}
	}
	return fireStoreUpdates
}

func (f *FirestoreHCShowCalendarService) DeleteShow(id string) error {
	_, err := f.database.Collection(utils.SHOW_COLLECTION).Doc(id).Delete(f.ctx)
	if err != nil {
		fmt.Println("error deleting show")
		fmt.Println(err)
		return errors.New("error deleting show")
	}
	return nil
}

func (f *FirestoreHCShowCalendarService) GetUser(username string) (*models.User, error) {
	dsnap, err := f.database.Collection(utils.USER_COLLECTION).Doc(username).Get(f.ctx)
	if err != nil {
		fmt.Println("error getting user")
		fmt.Println(err)
		return nil, errors.New("error getting user")
	}
	var u models.User
	err = dsnap.DataTo(&u)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("error getting user")
	}
	u.Password = ""
	return &u, nil
}

func (f *FirestoreHCShowCalendarService) GetUserShows(username string) (*[]models.Show, error) {
	var shows []models.Show
	iter := f.database.Collection(utils.SHOW_COLLECTION).Where("promoter", "==", username).OrderBy("date", firestore.Asc).Documents(f.ctx)
	for {
		s := models.Show{}
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println(err)
			fmt.Println("error iterating over results")
			return nil, err
		}
		err = doc.DataTo(&s)
		if err != nil {
			fmt.Println("error getting all shows")
			fmt.Println(err)
			return nil, errors.New("error getting all shows")
		}
		shows = append(shows, s)
	}
	return &shows, nil
}

func (f *FirestoreHCShowCalendarService) CreateUser(user models.User) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("error generating password hash")
		fmt.Println(err)
		return nil, err
	}

	u := models.User{Username: user.Username, Password: string(hashedPassword), Email: user.Email, Usertype: user.Usertype}
	_, err = f.database.Collection(utils.USER_COLLECTION).Doc(u.Username).Set(f.ctx, u)
	if err != nil {
		fmt.Println("some sort of error building the add query from firestore")
		fmt.Println(err)
		return nil, err
	}
	return &user, nil
}

func (f *FirestoreHCShowCalendarService) UpdateUser(username string, user models.User) (*models.User, error) {
	userFirestoreUpdateData := buildUserFirestoreUpdateData(user)
	_, err := f.database.Collection(utils.USER_COLLECTION).Doc(username).Update(f.ctx, userFirestoreUpdateData)
	if err != nil {
		fmt.Println("error updating user")
		fmt.Println(err)
		return nil, errors.New("error updating user")
	}

	return &user, nil
}

// TODO this whole function kinda sucks but works so make it better sometime
// is there a better way to get data from a request -> validate -> pass to update?
func buildUserFirestoreUpdateData(user models.User) []firestore.Update {
	fireStoreUpdates := []firestore.Update{}
	//i dont like this but i dont think i have a better choice to
	//turn my go struct into something firestore will like
	userTempGenericMap := structs.Map(user)
	for i, v := range userTempGenericMap {
		if strings.ToLower(i) != "password" {
			fireStoreUpdates = append(fireStoreUpdates, firestore.Update{Path: strings.ToLower(i), Value: v})
		}
	}
	return fireStoreUpdates
}

func (f *FirestoreHCShowCalendarService) ResetPassword(email string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("error generating password hash")
		fmt.Println(err)
		return err
	}

	var u models.User
	iter := f.database.Collection(utils.USER_COLLECTION).Where("email", "==", email).Documents(f.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = doc.DataTo(&u)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	if u.Username == "" {
		fmt.Println("user does not exist")
		return errors.New("user does not exist")
	}

	_, err = f.database.Collection(utils.USER_COLLECTION).Doc(u.Username).Update(f.ctx, []firestore.Update{{Path: "password", Value: string(hashedPassword)}})
	if err != nil {
		fmt.Println("error updating user")
		fmt.Println(err)
		return errors.New("error updating user")
	}

	return nil
}

func (f *FirestoreHCShowCalendarService) DeleteUser(username string) error {
	_, err := f.database.Collection(utils.USER_COLLECTION).Doc(username).Delete(f.ctx)
	if err != nil {
		fmt.Println("error deleting user")
		fmt.Println(err)
		return errors.New("error deleting user")
	}
	return nil
}

func (f *FirestoreHCShowCalendarService) AuthUser(user models.User) (string, error) {
	var u models.User
	dsnap, err := f.database.Collection(utils.USER_COLLECTION).Doc(user.Username).Get(f.ctx)
	if err != nil {
		fmt.Println("error getting user")
		fmt.Println(err)
		return "", utils.ErrUserDoesntExist
	}
	err = dsnap.DataTo(&u)
	if err != nil {
		fmt.Println(err)
		return "", utils.ErrUserDataMalformed
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(user.Password))
	if err != nil {
		fmt.Println("password does not match")
		fmt.Println(err)
		return "", utils.ErrUnauthorized
	}

	t, err := utils.GenerateToken(u.Username)
	if err != nil {
		fmt.Println("error generating access token")
		fmt.Println(err)
		return "", utils.ErrTokenGeneration
	}

	return t, nil
}

func (f *FirestoreHCShowCalendarService) CreateAuthObject(verification *models.Verification) error {
	_, err := f.database.Collection(utils.VERIFICATION_COLLECTION).Doc(verification.Email).Set(f.ctx, verification)
	if err != nil {
		fmt.Println("some sort of error building the add query from firestore")
		fmt.Println(err)
		return err
	}
	return nil
}

func (f *FirestoreHCShowCalendarService) DeleteAuthObject(email string) error {
	_, err := f.database.Collection(utils.VERIFICATION_COLLECTION).Doc(email).Delete(f.ctx)
	if err != nil {
		fmt.Println("error deleting verification object")
		fmt.Println(err)
		return errors.New("error deleting verification object")
	}
	return nil
}

func (f *FirestoreHCShowCalendarService) ValidateAuthUser(email string, code string) error {
	dsnap, err := f.database.Collection(utils.VERIFICATION_COLLECTION).Doc(email).Get(f.ctx)
	if err != nil {
		fmt.Println("error getting verification")
		fmt.Println(err)
		return errors.New("error getting verification")
	}
	var v models.Verification
	err = dsnap.DataTo(&v)
	if err != nil {
		fmt.Println(err)
		return errors.New("error getting verification")
	}
	if (code != v.Code) || time.Now().After(v.ExpiresAt) {
		fmt.Println("invalid verification")
		return errors.New("invalid verification")
	}
	if time.Now().After(v.ExpiresAt) {
		err := f.DeleteAuthObject(email)
		if err != nil {
			fmt.Println("error deleting validation object")
			return errors.New("invalid verification")
		}
		return errors.New("invalid verification")
	}

	return nil
}
