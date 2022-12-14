package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/leoashish99/bookings/pkg/config"
	"github.com/leoashish99/bookings/pkg/handlers"
	"github.com/leoashish99/bookings/pkg/render"
)

var portNumber = ":8080"
var app config.AppConfig
var session *scs.SessionManager

func main() {

	//Change this to true in Production
	//Variable shadowing
	app.InProduction = true
	//session settings
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session
	app.UseCache = false
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Cannot create template cache!!!")
	}
	app.TemplateCache = tc

	app.UseCache = true

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	render.NewTemplate(&app)
	fmt.Println(fmt.Sprintf("Starting application on port %s", portNumber))
	// http.ListenAndServe(portNumber, nil)
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	log.Fatal(err)
}

// func Routes(app *config.AppConfig) http.Handler {
// 	mux := pat.New()

// 	mux.Get("/", http.HandlerFunc(handlers.Repo.Home))
// 	mux.Get("/about", http.HandlerFunc(handlers.Repo.About))

// 	return mux
// }
