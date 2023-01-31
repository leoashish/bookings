package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/leoashish99/bookings/internal/config"
	"testing"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)

	switch v := mux.(type) {
	case *chi.Mux:
	default:
		t.Error(fmt.Sprintf("Not a correct Mux!!! It is %v", v))
	}
}
