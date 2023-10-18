package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/leoashish99/bookings/internal/driver"
	"github.com/leoashish99/bookings/internal/forms"
	"github.com/leoashish99/bookings/internal/helpers"
	"github.com/leoashish99/bookings/internal/repository"
	dbrepo "github.com/leoashish99/bookings/internal/repository/dbRepo"

	"github.com/leoashish99/bookings/internal/config"
	"github.com/leoashish99/bookings/internal/models"
	"github.com/leoashish99/bookings/internal/render"
)

var Repo *Repository

type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// New Handlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	m.App.Session.Put(r.Context(), "remoteIP", remoteIP)
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// renders the room page.
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)

	if err != nil {
		helpers.ServerError(w, err)
	}

	endDate, err := time.Parse(layout, end)

	if err != nil {
		helpers.ServerError(w, err)
	}
	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		return
	}

	if len(rooms) == 0 {
		// no Availability
		m.App.Session.Put(r.Context(), "error", "No Availability!!")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}
	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("Connot get the reservation from the session"))
	}

	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	roomData, err := m.DB.GetRoomById(res.RoomID)
	if err != nil {
		helpers.ServerError(w, err)
	}
	data := make(map[string]interface{})
	res.Room.RoomName = roomData.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	data["reservation"] = res
	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

// Post reservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	res := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	roomId, err := strconv.Atoi(r.Form.Get("room_id"))
	if res.EndDate.Format("2006-01-02") == "" {
		fmt.Println("End Date is NULL.")
	}
	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Phone:     r.Form.Get("phone"),
		Email:     r.Form.Get("email"),
		StartDate: res.StartDate,
		EndDate:   res.EndDate,
		RoomID:    roomId,
		Room:      res.Room,
	}

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email")
	form.MinLength("email", 3, r)
	form.IsEmail("email")
	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
	}
	newReservationId, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
	}

	roomRestriction := models.RoomRestrictions{
		StartDate:     res.StartDate,
		EndDate:       res.EndDate,
		RoomId:        roomId,
		ReservationId: newReservationId,
		RestrictionId: 1,
		Room:          models.Room{},
		Reservation:   models.Reservation{},
		Restriction:   models.Restriction{},
	}
	err = m.DB.InsertRoomRestrictions(roomRestriction)

	if err != nil {
		helpers.ServerError(w, err)
	}

	// Send e-mail notifications to the guest..
	htmlMessage := fmt.Sprintf(`
	<strong> Reservation Confirmation </strong> </br>
	Dear %s, </br>
	This is to confirm that your reservation for room %s from %s to %s.
   `, reservation.FirstName, reservation.Room.RoomName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg := models.MailData{
		To:       reservation.Email,
		From:     "leoashish99@gmail.com",
		Subject:  "Reservation Confirmation",
		Content:  htmlMessage,
		Template: "basic.html",
	}
	m.App.MailChan <- msg

	//Send an email notification to the owner
	htmlMessage = fmt.Sprintf(`
	<strong> Reservation Confirmation </strong> </br>
	Dear %s, </br>
	This is to confirm that your reservation for room %s from %s to %s.
   `, reservation.FirstName, reservation.Room.RoomName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg = models.MailData{
		To:      "leoashish99@gmail.com",
		From:    "leoashish99@gmail.com",
		Subject: "Reservation Confirmation",
		Content: htmlMessage,
	}
	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	RoomId    string `json:"room_id"`
}

func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	room_id, _ := strconv.Atoi(r.Form.Get("room_id"))

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	available, _ := m.DB.SearchAvailabilityByDatesByRoomId(startDate, endDate, room_id)
	fmt.Println(available)
	resp := jsonResponse{
		OK:        available,
		Message:   "Available!",
		StartDate: sd,
		EndDate:   ed,
		RoomId:    strconv.Itoa(room_id),
	}

	out, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application-json")
	w.Write(out)
	fmt.Println(resp.OK)
}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("Can't get the error from session.")
		m.App.Session.Put(r.Context(), "error", "Can't get the reservation for the session!!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
	m.App.Session.Remove(r.Context(), "reservation")
	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)

	stringMap["start_date"] = sd
	stringMap["end_date"] = ed
	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	if !ok {
		helpers.ServerError(w, err)
		return
	}
	res.RoomID = roomID
	m.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	ID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	sd := r.URL.Query().Get("sd")
	ed := r.URL.Query().Get("ed")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	roomData, err := m.DB.GetRoomById(ID)
	if err != nil {
		helpers.ServerError(w, err)
	}
	var room models.Room
	room.RoomName = roomData.RoomName
	room.ID = ID

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    ID,
		Room:      room,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// Logs in a user.
func (m *Repository) LoginPost(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
	}
	id, _, err := m.DB.Authenticate(email, password)

	if err != nil {
		log.Println(err)

		m.App.Session.Put(r.Context(), "error", "Invalid Login Credentials!!")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully!!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (*Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

// Shows all new reservations in the Admin Tool.
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.GetAllNewReservations()
	if err != nil {
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	fmt.Println(len(reservations))
	render.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// Shows all reservation
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.GetAllReservations()
	if err != nil {
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	fmt.Println(len(reservations))
	render.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}
func (m *Repository) AdminReservationsCalender(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	var current_year, current_month string

	if r.URL.Query().Get("y") != "" {
		current_year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		current_month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(current_year, time.Month(current_month), 1, 0, 0, 0, 0, time.UTC)
	}
	current_year = now.Format("2006")
	current_month = now.Format("01")

	nextMonthTime := now.AddDate(0, 1, 0)
	nextMonthYear := nextMonthTime.Format("2006")
	nextMonth := nextMonthTime.Format("01")

	previousMonthTime := now.AddDate(0, -1, 0)
	previousMonthYear := previousMonthTime.Format("2006")
	previousMonth := previousMonthTime.Format("01")

	stringMap := make(map[string]string)

	stringMap["current_year"] = current_year
	stringMap["current_month"] = current_month

	stringMap["next_month"] = nextMonth
	stringMap["next_year"] = nextMonthYear

	stringMap["previous_year"] = previousMonthYear
	stringMap["previous_month"] = previousMonth

	stringMap["current_month_string"] = now.Month().String()

	dataMap := make(map[string]interface{})
	dataMap["now"] = now

	// Get the number of days in the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := m.DB.AllRooms()

	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	dataMap["rooms"] = rooms
	render.Template(w, r, "admin-reservations-calender.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      dataMap,
		IntMap:    intMap,
	})
}

func (m *Repository) ShowReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParamFromCtx(r.Context(), "id"))
	processedString := chi.URLParamFromCtx(r.Context(), "src")

	stringMap := make(map[string]string)
	stringMap["src"] = processedString

	reservation, err := m.DB.GetReservation(id)
	if err != nil {
		return
	}
	data := make(map[string]interface{})
	fmt.Println(reservation.ID)
	data["reservation"] = reservation
	render.Template(w, r, "admin-show-reservation.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
		Form:      forms.New(nil),
	})
}

func (m *Repository) ShowReservationPOST(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		fmt.Println(err)
		return
	}

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	res, err := m.DB.GetReservation(id)

	if err != nil {
		fmt.Println(err)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Phone = r.Form.Get("phone")
	res.Email = r.Form.Get("email")

	err = m.DB.UpdateReservation(res)

	if err != nil {
		fmt.Println(err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Saved your reservation!!!")

	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}

func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	_ = m.DB.UpdateProcessedForReservation(id, 1)
	m.App.Session.Put(r.Context(), "flash", "Reservation Processed!!!")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)

}
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	_ = m.DB.DeleteReservation(id)
	m.App.Session.Put(r.Context(), "flash", "Reservation Deleted!!!")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)

}
