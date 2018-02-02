package store

import "time"

type Service interface {
	StoreEmail(user User) error
	ValidateEmail(user User) error
	DeleteEmail(user User) error
}

type User struct {
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	SignupDate time.Time `json:"signup"`
	Validated  bool      `json:"validated"`
}
