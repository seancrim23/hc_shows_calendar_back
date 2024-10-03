package models

import "time"

type Lineup []string

type Show struct {
	Id     string `json:"id" firestore:"id,omitempty"`
	Lineup `json:"lineup" firestore:"lineup,omitempty"`
	Date   time.Time `json:"date" firestore:"date,omitempty"`
	//only 2 character state code
	State string `json:"state" firestore:"state,omitempty"`
	City  string `json:"city" firestore:"city,omitempty"`
	Venue string `json:"venue" firestore:"venue,omitempty"`
	//street name number: 123 example st.
	Address string `json:"address" firestore:"address,omitempty"`
	//this should be the promoter id / username
	Promoter string `json:"promoter" firestore:"promoter,omitempty"`
}
