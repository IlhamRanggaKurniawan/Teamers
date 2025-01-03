package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/IlhamRanggaKurniawan/Teamers.git/internal/database/entity"
	"github.com/IlhamRanggaKurniawan/Teamers.git/internal/utils"
)

type Handler struct {
	userService UserService
}

type Input struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	ConfPassword string `json:"confPassword"`
}

type AuthRes struct {
	User        entity.User `json:"user"`
	AccessToken string      `json:"accessToken"`
}

func NewHandler(userService UserService) Handler {
	return Handler{
		userService: userService,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input Input

	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		utils.ErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	if input.Password != input.ConfPassword {
		utils.ErrorResponse(w, fmt.Errorf("password doesn't match"), http.StatusBadRequest)
		return
	}

	user, err := h.userService.Register(input.Username, input.Email, input.Password)

	if err != nil {
		utils.ErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	accessToken, err := utils.GenerateAndSetAccessToken(w, user.Id, user.Username, user.Email, user.Role)

	if err != nil {
		utils.ErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	_, err = utils.GenerateAndSetRefreshToken(w, user.Id, user.Username, user.Email, user.Role)

	if err != nil {
		utils.ErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	response := AuthRes{
		User:        *user,
		AccessToken: accessToken,
	}

	utils.SuccessResponse(w, response)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input Input

	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		utils.ErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	user, err := h.userService.Login(input.Email, input.Password)

	if err != nil {
		utils.ErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	accessToken, err := utils.GenerateAndSetAccessToken(w, user.Id, user.Username, user.Email, user.Role)

	if err != nil {
		utils.ErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	_, err = utils.GenerateAndSetRefreshToken(w, user.Id, user.Username, user.Email, user.Role)

	if err != nil {
		utils.ErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	response := AuthRes{
		User: *user,
		AccessToken: accessToken,
	}

	utils.SuccessResponse(w, response)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: "AccessToken",
		Value: "",
		Expires: time.Now().Add(-1),
		Secure: os.Getenv("APP_ENV") == "production",
		HttpOnly: true,
		Path: "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name: "RefreshToken",
		Value: "",
		Expires: time.Now().Add(-1),
		Secure: os.Getenv("APP_ENV") == "production",
		HttpOnly: true,
		Path: "/",
	})
	
	response := struct{
		Message string `json:"message"`
	}{
		Message: "Logout success",
	}

	utils.SuccessResponse(w, response)
}

func(h *Handler) GetToken(w http.ResponseWriter, r *http.Request) {
	user, err := utils.DecodeRefreshToken(r)

	if err != nil {
		utils.ErrorResponse(w, err, http.StatusUnauthorized)
		return
	}

	accessToken, err := utils.GenerateAndSetAccessToken(w, user.Id, user.Username, user.Email, user.Role)

	if err != nil {
		utils.ErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"accessToken": accessToken,
	}

	utils.SuccessResponse(w, response)
}
