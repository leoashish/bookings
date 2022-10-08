package render

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/leoashish99/bookings/pkg/config"
	"github.com/leoashish99/bookings/pkg/models"
)

var functions = template.FuncMap{}

var app *config.AppConfig

func addDefaultData(td *models.TemplateData) *models.TemplateData {

	return td
}

//New Template sets the config for the template package.
func NewTemplate(a *config.AppConfig) {
	app = a
}
func RenderTemplate(w http.ResponseWriter, tmpl string, td models.TemplateData) {
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
	pages, err := filepath.Glob("../../templates/*.page.html")

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

		matches, err := filepath.Glob("../../templates/*.layout.html")

		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("../../templates/*.layout.html")
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil

}
