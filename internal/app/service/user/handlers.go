package user

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"go-developer-course-diploma/internal/app/model"
	"go-developer-course-diploma/internal/app/storage"
	"net/http"
)

func WriteError(w http.ResponseWriter, code int, err error) {
	WriteResponse(w, code, err.Error())
}

func WriteResponse(w http.ResponseWriter, statusCode int, data string) {
	w.WriteHeader(statusCode)
	if len(data) != 0 {
		w.Write([]byte(data))
	}
}

func ReturnJSON(w http.ResponseWriter, statusCode int, body interface{}) {
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		WriteError(w, http.StatusInternalServerError, err)
		return
	}
	WriteResponse(w, statusCode, "")
}

func getPasswordHash(u *model.User) {
	u.Password = hex.EncodeToString([]byte(u.Password))
}

func RegisterHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user *model.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		if len(user.Login) == 0 || len(user.Password) == 0 {
			WriteError(w, http.StatusBadRequest, errors.New("login and password must NOT be empty"))
			return
		}

		getPasswordHash(user)

		err := s.User().RegisterUser(user)
		if err != nil && !errors.Is(err, storage.ErrorUserAlreadyExist) {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}
		if errors.Is(err, storage.ErrorUserAlreadyExist) {
			WriteError(w, http.StatusConflict, err)
			return
		}

		SetCookie(w, user.Login)
		WriteResponse(w, http.StatusOK, "")
	}
}

func LoginHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user *model.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			WriteError(w, http.StatusBadRequest, err)
			return
		}

		if len(user.Login) == 0 || len(user.Password) == 0 {
			WriteError(w, http.StatusBadRequest, errors.New("login and password must NOT be empty"))
			return
		}

		userDB, err := s.User().GetUser(user.Login)
		if err != nil && err != storage.ErrorUserNotFound {
			WriteError(w, http.StatusInternalServerError, err)
			return
		}
		if err == storage.ErrorUserNotFound {
			WriteError(w, http.StatusUnauthorized, err)
			return
		}

		getPasswordHash(user)

		if userDB.Login == user.Login && userDB.Password == user.Password {
			SetCookie(w, user.Login)
			WriteResponse(w, http.StatusOK, "")
			return
		}
		WriteResponse(w, http.StatusUnauthorized, "")
	}
}
