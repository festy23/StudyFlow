package handler

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	userpb "userservice/pkg/api"
)

//0. проходит через authmiddleware, до хэндлера в запросе уже есть x-user-id, x-user-role
//1. распарсить запрос: вытащить авторизацию в контекст, распарсить json в pb message
//2. отправить на клиент: запихнуть авторизацию в метадату, получить ответ, замапить и вернуть ошибку если есть
//3. распарсить ответ: распарсить pb message в json
//4. отправить ответ
//
//1. знать pb request, response
//2. принимать метод

type SignUpHandler struct {
	c userpb.UserServiceClient
}

func NewSignUpHandler(c userpb.UserServiceClient) *SignUpHandler {
	return &SignUpHandler{c: c}
}

func (h *SignUpHandler) SignUpViaTelegram(w http.ResponseWriter, r *http.Request) {
	handler, err := Handle[userpb.RegisterViaTelegramRequest, userpb.User](h.c.RegisterViaTelegram, nil, true)
	if err != nil {
		panic(err)
	}

	handler(w, r)
}

func (h *SignUpHandler) RegisterRoutes(r chi.Router) {
	r.Post("/sign-up/telegram", h.SignUpViaTelegram)
}
