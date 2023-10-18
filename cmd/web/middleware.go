package main

import (
	"fmt"
	"github.com/leoashish99/bookings/internal/helpers"
	"net/http"

	"github.com/justinas/nosurf"
)

func WriteToConsole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hit the page!!")
		next.ServeHTTP(w, r)
	})
}

func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// load and saves the session on every request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			session.Put(r.Context(), "error", "Please login first!!!")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
