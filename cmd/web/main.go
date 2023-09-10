package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/leoashish99/bookings/internal/driver"
	"github.com/leoashish99/bookings/internal/helpers"
	"github.com/leoashish99/bookings/internal/models"

	"github.com/alexedwards/scs/v2"
	"github.com/leoashish99/bookings/internal/config"
	"github.com/leoashish99/bookings/internal/handlers"
	"github.com/leoashish99/bookings/internal/render"
)

var portNumber = "localhost:8080"
var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func main() {
	db, error := run()
	if error != nil {
		log.Fatal(error)
	}
	defer db.SQL.Close()

	fmt.Println(fmt.Sprintf("Starting application on port %s", portNumber))
	// http.ListenAndServe(portNumber, nil)
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err := srv.ListenAndServe()
	log.Fatal(err)
}
func run() (*driver.DB, error) {
	//What I am going to put in the session
	gob.Register(models.Reservation{})
	gob.Register(models.Restriction{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	//Change this to true in Production
	//Variable shadowing
	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog
	//session settings
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	log.Println("Connecting to the database...")
	//Connecting to database
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=bookings user=postgres password=singoo123#")
	if err != nil {
		log.Fatal("Cannot connect to the database! Dying....")
		return nil, err
	}
	log.Println("Connected to the database!!!")

	app.UseCache = false
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Cannot create template cache!!!")
		return nil, err
	}
	app.TemplateCache = tc

	app.UseCache = true

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	helpers.NewHelper(&app)
	render.NewRenderer(&app)
	return db, err
}
