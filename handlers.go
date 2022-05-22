package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"net/http"
)

var buf = make(chan []byte, 64)

func Run() {
	pool := make(chan []byte, 64)
	i := 1
	for {
		select {
		case output := <-buf:
			pool <- output
			if i <= 64 {
				go work(pool)
				i++
			}
		}
	}
}

func work(pool <-chan []byte) {
	for msg := range pool {
		Send(string(msg))
	}
}

func getCakeHandler(w http.ResponseWriter, r *http.Request, u User, users UserRepository) {
	w.Write([]byte(u.FavoriteCake))
	buf <- []byte("Someone got cake")
	GivenCakesCount.Inc()
}

func getMeHandler(w http.ResponseWriter, r *http.Request, u User, users UserRepository) {
	w.Write([]byte(u.FavoriteCake))
	w.Write([]byte(u.Email))
	buf <- []byte("Someone got himself")
	UserInfoCount.Inc()
}

func getEmailHandler(w http.ResponseWriter, r *http.Request, u User, users UserRepository) {
	w.Write([]byte(u.Email))
	buf <- []byte("Someone got email")
}

func (uServ UserService) updateCakeHandler(w http.ResponseWriter, r *http.Request, u User, users UserRepository) {
	params := &UserRegisterParams{}
	err := json.NewDecoder(r.Body).Decode(params)

	if err != nil {
		handleError(errors.New("could not read params"), w)
		return
	}

	if err := validateRegisterParams(params); err != nil {
		handleError(err, w)
		return
	}

	passwordDigest := md5.New().Sum([]byte(params.Password))

	newCake := User{
		Email:          params.Email,
		PasswordDigest: string(passwordDigest),
		FavoriteCake:   params.FavoriteCake,
	}

	err = uServ.repository.Update(params.Email, newCake)
	if err != nil {
		handleError(err, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("cake updated"))
	buf <- []byte("Someone updated cake")
}

func (uServ UserService) updateEmailHandler(w http.ResponseWriter, r *http.Request, u User, users UserRepository) {
	params := &UserRegisterParams{}
	err := json.NewDecoder(r.Body).Decode(params)

	if err != nil {
		handleError(errors.New("could not read params"), w)
		return
	}

	if err := validateRegisterParams(params); err != nil {
		handleError(err, w)
		return
	}

	passwordDigest := md5.New().Sum([]byte(params.Password))
	email := params.Email

	newEmail := User{
		Email:          email,
		PasswordDigest: string(passwordDigest),
		FavoriteCake:   params.FavoriteCake,
	}

	uServ.repository.Delete(u.Email)
	err = uServ.repository.Add(email, newEmail)

	if err != nil {
		handleError(err, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("email updated"))
	buf <- []byte("Someone updated email")
}

func (uServ UserService) updatePasswordHandler(w http.ResponseWriter, r *http.Request, u User, users UserRepository) {
	params := &UserRegisterParams{}
	err := json.NewDecoder(r.Body).Decode(params)

	if err != nil {
		handleError(errors.New("could not read params"), w)
		return
	}

	if err := validateRegisterParams(params); err != nil {
		handleError(err, w)
		return
	}

	passwordDigest := md5.New().Sum([]byte(params.Password))
	newCake := User{
		Email:          params.Email,
		PasswordDigest: string(passwordDigest),
		FavoriteCake:   params.FavoriteCake,
	}

	err = uServ.repository.Update(params.Email, newCake)
	if err != nil {
		handleError(err, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("password updated"))
	buf <- []byte("Someone updated password")
}
