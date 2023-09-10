package repository

import (
	"github.com/leoashish99/bookings/internal/models"
	"time"
)

type DatabaseRepo interface {
	AllUsers() bool

	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestrictions(res models.RoomRestrictions) error
	SearchAvailabilityByDates(startDate, endDate time.Time, roomId int) (bool, error)
}
