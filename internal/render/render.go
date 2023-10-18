package render

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/justinas/nosurf"
	"github.com/leoashish99/bookings/internal/config"
	"github.com/leoashish99/bookings/internal/models"
)

var functions = template.FuncMap{
	"humanDate":  HumanDate,
	"formatDate": FormatDate,
	"iterate":    Iterate,
	"add":        Add,
}

var app *config.AppConfig

func Add(a, b int) int {
	return a + b
}
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = true
	}
	return td
}

func Iterate(count int) []int {
	var i int
	var items []int
	for i = 0; i < count; i++ {
		items = append(items, i)
	}
	return items
}

// New Template sets the config for the template package.
func NewRenderer(a *config.AppConfig) {
	app = a
}

// returns time in yyyy-mm-dd format
func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}
func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) {
	var tc map[string]*template.Template
	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	tc = app.TemplateCache
	t, ok := tc[tmpl]

	if !ok {
		log.Fatal("Could not get template from template cache.")
	}

	buf := new(bytes.Buffer)
	td = AddDefaultData(td, r)
	_ = t.Execute(buf, td)

	_, err := buf.WriteTo(w)

	if err != nil {
		fmt.Println("Error writing template to the browser.")
	}
}

// creates a template cache.
func CreateTemplateCache() (map[string]*template.Template, error) {
	fmt.Println("Here")
	myCache := map[string]*template.Template{}
	pages, err := filepath.Glob("../../templates/*.page.tmpl")

	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		fmt.Println("Page is currently ", page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)

		if err != nil {
			return myCache, err

		}

		matches, err := filepath.Glob("../../templates/*.layout.tmpl")

		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("../../templates/*.layout.tmpl")
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}
	fmt.Printf("Successfully creating the Cache!!")
	return myCache, nil

}
