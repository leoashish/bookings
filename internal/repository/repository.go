package repository

import (
	"github.com/leoashish99/bookings/internal/models"
	"time"
)

type DatabaseRepo interface {
	AllUsers() bool

	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestrictions(res models.RoomRestrictions) error
	SearchAvailabilityByDatesByRoomId(startDate, endDate time.Time, roomId int) (bool, error)
	SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error)
	GetRoomById(id int) (models.Room, error)
	GetUserById(ID int) (models.User, error)
	Authenticate(email, testPassword string) (int, string, error)
	GetAllReservations() ([]models.Reservation, error)
	GetAllNewReservations() ([]models.Reservation, error)
	GetReservation(id int) (models.Reservation, error)
	UpdateReservation(reservation models.Reservation) error
	DeleteReservation(id int) error
	UpdateProcessedForReservation(id, processed int) error
	AllRooms() ([]models.Room, error)
}
