package config

import (
	"html/template"

	"github.com/alexedwards/scs/v2"
)

// whenever you do feel like doing start doing and it will complete
// Have a strong belief that you can

//AppConfig  holds the application config
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InProduction  bool
	Session       *scs.SessionManager
}
