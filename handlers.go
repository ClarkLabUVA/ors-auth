package main

import (
  "net/http"
)




// POST /user/
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {}


// GET /user/
func ListUserHandler(w http.ResponseWriter, r *http.Request) {}


// GET /user/:userID
func GetUserHandler(w http.ResponseWriter, r *http.Request) {}


// PUT /user/:userID
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {}


// Delete /user/:userID
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {}
