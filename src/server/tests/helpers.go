package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/matthewdeguzman/GatorGuessr/src/server/endpoints"
	"github.com/matthewdeguzman/GatorGuessr/src/server/endpoints/api"
	u "github.com/matthewdeguzman/GatorGuessr/src/server/structs"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func testInitMigration(t *testing.T) (db *gorm.DB) {
	const DB_USERNAME = "cen3031"
	const DB_NAME = "test_database"
	const DB_HOST = "cen3031-server.mysql.database.azure.com"
	const DB_PORT = "3306"

	// Build connection string
	DSN := DB_USERNAME + ":" + DB_PASSWORD + "@tcp" + "(" + DB_HOST + ":" + DB_PORT + ")/" + DB_NAME + "?" + "parseTime=true&loc=Local"

	db, err := gorm.Open(mysql.Open(DSN), &gorm.Config{})
	if err != nil {
		fmt.Println(err.Error())
		panic("Cannot connect to DB")
	}

	// migrates the server if necessary
	db.AutoMigrate(&u.User{})

	return db
}

func cleanDB(user *u.User, username string, t *testing.T) {
	db := testInitMigration(t)
	db.Delete(user, "Username = ?", user.Username)
}

func addUser(user u.User, t *testing.T) (err error) {
	db := testInitMigration(t)
	hash, err := endpoints.EncodePassword(user.Password)

	if err != nil {
		t.Error(err)
	}
	user.Password = hash

	db.Create(&user)
	return nil
}

/// MOCK FUNCTIONS ///

func mockGetUsers(w http.ResponseWriter, r *http.Request, t *testing.T) {
	db := testInitMigration(t)
	api.GetUsers(w, r, db)
}

func mockGetUser(w http.ResponseWriter, r *http.Request, username string, t *testing.T) {
	db := testInitMigration(t)
	api.GetUserWithUsername(w, r, username, db)
}

func mockCreateUser(w http.ResponseWriter, r *http.Request, user u.User, db *gorm.DB, t *testing.T) {
	api.CreateUserFromUser(w, r, user, db)
}

func mockUpdateUser(w http.ResponseWriter, r *http.Request, oldUser, updatedUser u.User, db *gorm.DB, t *testing.T) {
	UpdateUserFromUser(w, r, username, db)
}

func mockDeleteUser(w http.ResponseWriter, r *http.Request, username string, db *gorm.DB, t *testing.T) {
	endpoints.SetHeader(w)
	var user u.User

	endpoints.FetchUser(db, &user, username)
	if user.Username == "" {
		endpoints.UserDNErr(w)
		return
	}
	db.Delete(&user, "Username = ?", username)
	endpoints.EncodeUser(user, w)
}

func mockValidateUser(w http.ResponseWriter, r *http.Request, user u.User, db *gorm.DB, t *testing.T) {
	endpoints.SetHeader(w)

	var givenPassword string
	var hashedPassword string
	var dbUser u.User
	givenPassword = user.Password

	endpoints.FetchUser(db, &dbUser, user.Username)

	if dbUser.Username == "" {
		endpoints.UserDNErr(w)
		return
	}
	hashedPassword = dbUser.Password

	match, err := endpoints.DecodePasswordAndMatch(givenPassword, hashedPassword)
	if err != nil {
		t.Log("Hashed: " + hashedPassword)
		t.Log(err)
		endpoints.HashErr(w)
		return
	}
	if !match {
		endpoints.LoginErr(w)
		return
	}
}

func mockGetTopUsers(w http.ResponseWriter, r *http.Request, limit string, db *gorm.DB, t *testing.T) {
	var users []u.User

	lim, err := strconv.Atoi(limit)
	if err != nil {
		endpoints.WriteErr(w, http.StatusBadRequest, "400 - Could not process limit")
		return
	}
	if lim <= 0 {
		endpoints.WriteErr(w, http.StatusBadRequest, "400 - Limit must be a positive integer")
		return
	}

	db.Limit(lim).Order("score desc").Find(&users)
	endpoints.EncodeUsers(users, w)
}

// TESTING FUNCTIONS //

func getUserTest(username string, t *testing.T) (status int) {
	req, err := http.NewRequest("GET", "/api/users/{username}/", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockGetUser(w, r, username, t)
	})

	handler.ServeHTTP(rr, req)

	return rr.Result().StatusCode
}

func createUserTest(user u.User, t *testing.T) (status int) {
	db := testInitMigration(t)
	req, err := http.NewRequest("POST", "/api/users/", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockCreateUser(w, r, user, db, t)
	})

	handler.ServeHTTP(rr, req)

	return rr.Code
}

func updateUserTest(user map[string]string, username string, t *testing.T) (status int) {
	db := testInitMigration(t)
	req, err := http.NewRequest("PUT", "/api/users/{username}/", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Write()

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockUpdateUser(w, r, user, username, db, t)
	})

	handler.ServeHTTP(rr, req)

	return rr.Code
}

func deleteUserTest(username string, t *testing.T) (status int) {
	db := testInitMigration(t)
	req, err := http.NewRequest("DELETE", "/api/users/{username}/", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockDeleteUser(w, r, username, db, t)
	})

	handler.ServeHTTP(rr, req)

	return rr.Code
}

func validateUserTest(user u.User, t *testing.T) (status int) {
	db := testInitMigration(t)
	req, err := http.NewRequest("POST", "/api/login/", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockValidateUser(w, r, user, db, t)
	})

	handler.ServeHTTP(rr, req)

	return rr.Code
}

func getTopUsersTest(limit string, t *testing.T) (status int, users []u.User) {
	db := testInitMigration(t)
	req, err := http.NewRequest("GET", "/api/leaderboard/{limit}/", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockGetTopUsers(w, r, limit, db, t)
	})

	handler.ServeHTTP(rr, req)

	status = rr.Code
	json.NewDecoder(rr.Result().Body).Decode(&users)

	return status, users
}
