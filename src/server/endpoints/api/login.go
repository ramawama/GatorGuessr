package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	helpers "github.com/matthewdeguzman/GatorGuessr/src/server/endpoints"
	cookies "github.com/matthewdeguzman/GatorGuessr/src/server/endpoints/cookies"
	u "github.com/matthewdeguzman/GatorGuessr/src/server/structs"

	"gorm.io/gorm"
)

func GetUserWithUsername(w http.ResponseWriter, r *http.Request, username string, db *gorm.DB) {
	var user u.User
	helpers.FetchUser(db, &user, username)

	if user.Username == "" {
		helpers.UserDNErr(w)
		return
	}
	err := helpers.AuthorizeRequest(w, r, user)
	if err != nil {
		http.Error(w, "Access Denied", http.StatusForbidden)
		return
	}
	helpers.EncodeUser(user, w)
}

func CreateUserFromUser(w http.ResponseWriter, r *http.Request, user u.User, db *gorm.db) {
	if helpers.UserExists(db, user.Username) {
		helpers.WriteErr(w, http.StatusBadRequest, "400 - User already exists")
		return
	}

	if user.ID != 0 || user.Password == "" {
		helpers.WriteErr(w, http.StatusBadRequest, "400 - Attempting to change ID or password is empty")
		return
	}

	hash, err := helpers.EncodePassword(user.Password)

	if err != nil {
		helpers.HashErr(w)
		return
	}
	user.Password = hash

	db.Create(&user)
	helpers.EncodeUser(user, w)

	// create cookie
	cookies.SetCookieHandler(w, r, user)
}

func UpdateUserFromUser(w http.ResponseWriter, r *http.Request) {
	if oldUser.Username == "" {
		helpers.UserDNErr(w)
		return
	}

	// validate request
	err := helpers.AuthorizeRequest(w, r, oldUser)
	if err != nil {
		http.Error(w, "Access Denied", http.StatusForbidden)
		return
	}

	helpers.DecodeUser(&updatedUser, r)

	if oldUser.ID != updatedUser.ID {
		helpers.WriteErr(w, http.StatusMethodNotAllowed, "405 - Cannot change immutable field")
		return
	}

	hash, err := helpers.EncodePassword(updatedUser.Password)
	if err != nil {
		helpers.HashErr(w)
		return
	}
	updatedUser.Password = hash
	updatedUser.CreatedAt = oldUser.CreatedAt

	db.Save(&updatedUser)
	helpers.EncodeUser(updatedUser, w)
}
func GetUsers(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	helpers.SetHeader(w)
	var users []u.User
	db.Find(&users)
	json.NewEncoder(w).Encode(users)
}

func GetUser(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	helpers.SetHeader(w)

	params := mux.Vars(r)
	GetUserWithUsername(w, r, params["username"], db)
}

func CreateUser(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	helpers.SetHeader(w)

	var user u.User
	helpers.DecodeUser(&user, r)

	CreateUserFromUser(w, r, user, db)
}

func UpdateUser(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	helpers.SetHeader(w)
	params := mux.Vars(r)

	var oldUser u.User
	var updatedUser u.User

	helpers.FetchUser(db, &oldUser, params["username"])
	helpers.FetchUser(db, &updatedUser, params["username"])

	UpdateUserFromuser()
}

func DeleteUser(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	helpers.SetHeader(w)
	params := mux.Vars(r)
	var user u.User

	helpers.FetchUser(db, &user, params["username"])
	if user.Username == "" {
		helpers.UserDNErr(w)
		return
	}

	// authorize request
	err := helpers.AuthorizeRequest(w, r, user)
	if err != nil {
		http.Error(w, "Access Denied", http.StatusForbidden)
		return
	}
	db.Delete(&user, "Username = ?", params["username"])
	helpers.EncodeUser(user, w)
}

func ValidateUser(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	helpers.SetHeader(w)

	var user u.User
	var givenPassword string
	var hashedPassword string

	helpers.DecodeUser(&user, r)
	givenPassword = user.Password
	helpers.FetchUser(db, &user, user.Username)

	if user.Username == "" {
		helpers.UserDNErr(w)
		return
	}
	hashedPassword = user.Password

	match, err := helpers.DecodePasswordAndMatch(givenPassword, hashedPassword)
	if err != nil {
		helpers.HashErr(w)
		return
	}
	if !match {
		helpers.LoginErr(w)
		return
	}

	// create cookie
	cookies.SetCookieHandler(w, r, user)
}
