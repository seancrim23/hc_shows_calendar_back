package models

// TODO figure out what needs to be built out for users
type User struct {
	//username will be unique and serve as id
	Username string `json:"username" firestore:"username"`
	//Hash     string `json:"hash" firestore:"hash"`
	Email string `json:"email" firestore:"email"`
	//plain text should never exist in a DB
	Password string `json:"password" firestore:"password,omitempty"`
	Usertype string `json:"usertype" firestore:"usertype"`
}
