package main

import (
	"encoding/gob"
	"fmt"
	"github.com/leoashish99/bookings/internal/models"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/leoashish99/bookings/internal/config"
	"github.com/leoashish99/bookings/internal/handlers"
	"github.com/leoashish99/bookings/internal/render"
)

var portNumber = "localhost:8080"
var app config.AppConfig
var session *scs.SessionManager

func main() {
	error := run()
	if error != nil {
		log.Fatal(error)
	}

	fmt.Println(fmt.Sprintf("Starting application on port %s", portNumber))
	// http.ListenAndServe(portNumber, nil)
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err := srv.ListenAndServe()
	log.Fatal(err)
}
func run() error {
	//What I am going to put in the session
	gob.Register(models.Reservation{})
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
		return err
	}
	app.TemplateCache = tc

	app.UseCache = true

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	render.NewTemplate(&app)
	return err
}
