package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"hc_shows_backend/models"
	"hc_shows_backend/services"
	"hc_shows_backend/utils"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type HCShowCalendarServer struct {
	service      services.HCShowCalendarService
	emailService services.HCShowCalendarEmailService
	http.Handler
}

func NewHCShowCalendarServer(service services.HCShowCalendarService, emailService services.HCShowCalendarEmailService) (*HCShowCalendarServer, error) {
	h := new(HCShowCalendarServer)

	h.service = service
	h.emailService = emailService

	r := mux.NewRouter()
	r.HandleFunc("/health", h.healthCheck).Methods("GET")

	//TODO look into cleaning up endpoints
	r.HandleFunc("/show", h.getShows).Methods("GET")
	r.HandleFunc("/show", utils.WithToken(h.createShow)).Methods("POST")
	r.HandleFunc("/show/{id}", h.getShow).Methods("GET")
	r.HandleFunc("/show/{id}", utils.WithToken(h.updateShow)).Methods("PUT")
	r.HandleFunc("/show/{id}", utils.WithToken(h.deleteShow)).Methods("DELETE")

	r.HandleFunc("/auth", h.authUser).Methods("POST")
	//TODO add admin middleware
	r.HandleFunc("/auth/setup", utils.WithToken(h.authSetup)).Methods("POST")
	r.HandleFunc("/auth/reset", h.authReset).Methods("POST")

	r.HandleFunc("/user", h.createUser).Methods("POST")              //token
	r.HandleFunc("/user", utils.WithToken(h.getUser)).Methods("GET") //token
	r.HandleFunc("/user/shows", utils.WithToken(h.getUserShows)).Methods("GET")
	r.HandleFunc("/user", utils.WithToken(h.updateUser)).Methods("PUT")    // token
	r.HandleFunc("/user", utils.WithToken(h.deleteUser)).Methods("DELETE") //token

	//TODO TOKEN PROTECT
	//UPDATE RESET USER TO TAKE USER ID FROM CONTEXT AND USE TO UPDATE PASSWORD
	r.HandleFunc("/user/reset", h.resetUser).Methods("PUT")

	handler := cors.New(cors.Options{
		AllowedOrigins: []string{os.Getenv(utils.ALLOWED_ORIGINS)},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		Debug:          false,
	}).Handler(r)

	h.Handler = handler

	return h, nil
}

func (h *HCShowCalendarServer) getShow(w http.ResponseWriter, r *http.Request) {
	var code = 200
	var err error

	showId := mux.Vars(r)["id"]
	//TODO perform input validation
	if showId == "" {
		code = 400
		utils.RespondWithError(w, code, errors.New("no id passed to request").Error())
		return
	}

	show, err := h.service.GetShow(showId)
	//determine what type of error and change code and return according error message
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}
	if show == nil {
		code = 404
		utils.RespondWithError(w, code, errors.New("cannot find show").Error())
		return
	}

	utils.RespondWithJSON(w, code, show)
}

func (h *HCShowCalendarServer) getShows(w http.ResponseWriter, r *http.Request) {
	var code = 200
	var err error
	showQueryFilters := make(map[string]string)
	params := r.URL.Query()

	if len(params) > 0 {
		for k, v := range params {
			if _, ok := utils.ValidShowQueryFilters[strings.ToLower(k)]; ok {
				showQueryFilters[k] = v[0]
			}
		}
	}

	shows, err := h.service.GetShows(showQueryFilters)
	//TODO build out something to parse errors and deliver error codes
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}

	utils.RespondWithJSON(w, code, shows)
}

// user needs to be validated and username should get passed to show that is created
func (h *HCShowCalendarServer) createShow(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var code = 201
	var err error
	var show models.Show

	userID := r.Context().Value(utils.UserIDKey{}).(string)
	if userID == "" {
		code = 400
		fmt.Println("no user id provided")
		utils.RespondWithError(w, code, errors.New("no id passed to request").Error())
		return
	}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		fmt.Println(err)
		utils.RespondWithError(w, code, err.Error())
		return
	}

	err = json.Unmarshal(reqBody, &show)
	if err != nil {
		code = 400
		fmt.Println(err)
		utils.RespondWithError(w, code, err.Error())
		return
	}

	s, err := h.service.CreateShow(show, userID)
	if err != nil {
		code = 500
		fmt.Println(err)
		utils.RespondWithError(w, code, err.Error())
		return
	}

	utils.RespondWithJSON(w, code, s)
}

// TODO add validation so logged in promoter can only change their own shows?
func (h *HCShowCalendarServer) updateShow(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var code = 200
	var err error
	var show models.Show

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}
	err = json.Unmarshal(reqBody, &show)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}

	showId := mux.Vars(r)["id"]
	if showId == "" {
		code = 400
		utils.RespondWithError(w, code, errors.New("no id passed to request").Error())
		return
	}
	//perform input validation

	s, err := h.service.UpdateShow(showId, show)
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}

	utils.RespondWithJSON(w, code, s)
}

// TODO add validation so logged in promoter can only change their own shows?
func (h *HCShowCalendarServer) deleteShow(w http.ResponseWriter, r *http.Request) {
	var code = 200
	var err error

	showId := mux.Vars(r)["id"]
	if showId == "" {
		code = 400
		utils.RespondWithError(w, code, errors.New("no id passed to request").Error())
		return
	}
	//perform input validation

	err = h.service.DeleteShow(showId)
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}

	utils.RespondWithJSON(w, code, map[string]string{"response": "deleted"})
}

// any create user call should have the verification code on it...
// get user info
// get code
func (h *HCShowCalendarServer) createUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var code = 201
	var err error
	var user models.User
	var verification models.Verification

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}
	err = json.Unmarshal(reqBody, &user)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}

	//TODO see if i can do the code verification better
	err = json.Unmarshal(reqBody, &verification)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}

	err = h.service.ValidateAuthUser(user.Email, verification.Code)
	if err != nil {
		code = 400
		fmt.Println("user cannot be created...")
		utils.RespondWithError(w, code, err.Error())
		return
	}

	//TODO set all users with accounts up as promoters / possibly expand in future
	user.Usertype = "promoter"

	u, err := h.service.CreateUser(user)
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}

	err = h.service.DeleteAuthObject(user.Email)
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}

	utils.RespondWithJSON(w, code, u)
}

func (h *HCShowCalendarServer) resetUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var code = 201
	var err error
	var user models.User
	var verification models.Verification

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}
	err = json.Unmarshal(reqBody, &user)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}

	//TODO see if i can do the code verification better
	err = json.Unmarshal(reqBody, &verification)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}

	err = h.service.ValidateAuthUser(user.Email, verification.Code)
	if err != nil {
		code = 400
		fmt.Println("user cannot reset password...")
		utils.RespondWithError(w, code, err.Error())
		return
	}

	err = h.service.ResetPassword(user.Email, user.Password)
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}

	err = h.service.DeleteAuthObject(user.Email)
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}

	utils.RespondWithJSON(w, code, map[string]string{"response": "reset"})
}

func (h *HCShowCalendarServer) authUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var code = 200
	var err error
	var u models.User

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}
	err = json.Unmarshal(reqBody, &u)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}
	t, err := h.service.AuthUser(u)
	if err != nil {
		switch err {
		case utils.ErrTokenGeneration, utils.ErrUserDataMalformed:
			code = 400
		case utils.ErrUnauthorized:
			code = 401
		case utils.ErrUserDoesntExist:
			code = 404
		default:
			code = 500
		}
		utils.RespondWithError(w, code, err.Error())
		return
	}

	utils.RespondWithJSON(w, code, map[string]string{"token": t})
}

// TODO add some sort of admin token? like admin is only one allowed to do this?
func (h *HCShowCalendarServer) authSetup(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var code = 201
	var err error
	var v models.Verification

	userID := r.Context().Value(utils.UserIDKey{}).(string)
	if userID == "" {
		code = 400
		fmt.Println("no user id provided")
		utils.RespondWithError(w, code, errors.New("no id passed to request").Error())
		return
	}

	//TODO i think this is the lazy way to do this (even though its not terrible imo), update at some point to better way
	//maybe middleware function
	uObject, err := h.service.GetUser(userID)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}
	if uObject.Usertype != utils.ADMIN_USER {
		code = 400
		fmt.Println("user cannot setup new authentication")
		utils.RespondWithError(w, code, errors.New("user cannot setup new authentication").Error())
		return
	}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}
	err = json.Unmarshal(reqBody, &v)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}

	verificationCode := utils.GenerateRandomString(8)
	v.Code = verificationCode
	v.Type = models.MailConfirmation
	v.ExpiresAt = time.Now().Add(time.Hour * time.Duration(24))

	err = h.service.CreateAuthObject(&v)
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}

	from := utils.MY_EMAIL
	to := []string{v.Email}
	subject := "Email Verification for HC Shows Calendar"
	mailType := services.MailConfirmation
	mailData := &services.MailData{
		Email: v.Email,
		Code:  verificationCode,
	}

	mailReq := h.emailService.NewMail(from, to, subject, mailType, mailData)
	err = h.emailService.SendMail(mailReq)
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}

	utils.RespondWithJSON(w, code, map[string]string{"response": "created"})
}

func (h *HCShowCalendarServer) authReset(w http.ResponseWriter, r *http.Request) {
	//all you should need to pass this is an email
	defer r.Body.Close()
	var code = 201
	var err error
	var v models.Verification

	//should contain an email to setup the auth for
	//should create the verification object and store in the db
	//should send the account setup email

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}
	err = json.Unmarshal(reqBody, &v)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}

	verificationCode := utils.GenerateRandomString(8)
	v.Code = verificationCode
	v.Type = models.PassReset
	v.ExpiresAt = time.Now().Add(time.Hour * time.Duration(24))

	err = h.service.CreateAuthObject(&v)
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}

	from := utils.MY_EMAIL
	to := []string{v.Email}
	subject := "Password Reset for HC Shows Calendar"
	mailType := services.PassReset
	mailData := &services.MailData{
		Email: v.Email,
		Code:  verificationCode,
	}
	//should create an email for the particular type and then send
	mailReq := h.emailService.NewMail(from, to, subject, mailType, mailData)
	err = h.emailService.SendMail(mailReq)
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}

	utils.RespondWithJSON(w, code, map[string]string{"response": "created"})
}

func (h *HCShowCalendarServer) getUser(w http.ResponseWriter, r *http.Request) {
	var code = 200
	var err error

	userID := r.Context().Value(utils.UserIDKey{}).(string)
	if userID == "" {
		code = 400
		fmt.Println("no user id provided")
		utils.RespondWithError(w, code, errors.New("no id passed to request").Error())
		return
	}

	user, err := h.service.GetUser(userID)
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}
	if user == nil {
		code = 404
		utils.RespondWithError(w, code, errors.New("cannot find user").Error())
		return
	}

	utils.RespondWithJSON(w, code, user)
}

func (h *HCShowCalendarServer) getUserShows(w http.ResponseWriter, r *http.Request) {
	var code = 200
	var err error

	userID := r.Context().Value(utils.UserIDKey{}).(string)
	if userID == "" {
		code = 400
		fmt.Println("no user id provided")
		utils.RespondWithError(w, code, errors.New("no id passed to request").Error())
		return
	}

	shows, err := h.service.GetUserShows(userID)
	//TODO build out something to parse errors and deliver error codes
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}

	utils.RespondWithJSON(w, code, shows)
}

func (h *HCShowCalendarServer) updateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var code = 200
	var err error
	var user models.User

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}
	err = json.Unmarshal(reqBody, &user)
	if err != nil {
		code = 400
		utils.RespondWithError(w, code, err.Error())
		return
	}

	userID := r.Context().Value(utils.UserIDKey{}).(string)
	if userID == "" {
		code = 400
		fmt.Println("no user id provided")
		utils.RespondWithError(w, code, errors.New("no id passed to request").Error())
		return
	}

	u, err := h.service.UpdateUser(userID, user)
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}

	utils.RespondWithJSON(w, code, u)
}

func (h *HCShowCalendarServer) deleteUser(w http.ResponseWriter, r *http.Request) {
	var response interface{}
	var code = 200
	var err error

	userID := r.Context().Value(utils.UserIDKey{}).(string)
	if userID == "" {
		code = 400
		fmt.Println("no user id provided")
		utils.RespondWithError(w, code, errors.New("no id passed to request").Error())
		return
	}

	err = h.service.DeleteUser(userID)
	if err != nil {
		code = 500
		utils.RespondWithError(w, code, err.Error())
		return
	}

	utils.RespondWithJSON(w, code, response)
}

func (h *HCShowCalendarServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	//TODO research correct health check implementation
	io.WriteString(w, `{"alive": true}`)
}
