package dbrepo

import (
	"context"
	"github.com/leoashish99/bookings/internal/models"
	"time"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

// Inserts the reservation into the database
func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var newId int
	stmt := `insert into reservations(first_name, last_name,email,phone,
		start_date, end_date, room_id, created_at, updated_at)
		values($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`
	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newId)
	if err != nil {
		return 0, err
	}
	return newId, nil
}

// / Insert Room Restriction : inserts a room restriction into the
// / database
func (m *postgresDBRepo) InsertRoomRestrictions(res models.RoomRestrictions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stmt := `insert into room_restrictions(start_date,end_date,room_id,
             reservation_id,created_at, updated_at, restriction_id)
		values($1, $2, $3, $4, $5, $6, $7) returning id`
	_, err := m.DB.ExecContext(ctx, stmt,
		res.StartDate,
		res.EndDate,
		res.RoomId,
		res.ReservationId,
		time.Now(),
		time.Now(),
		res.RestrictionId,
	)
	if err != nil {
		return err
	}

	return nil
}

// True if the availability Exists elsewise false
func (m *postgresDBRepo) SearchAvailabilityByDates(startDate, endDate time.Time, roomId int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select count(*)
              from room_restrictions
              where end_date > $1 and start_date < $2
              and roomid = $3`

	var numOfRows int
	rows := m.DB.QueryRowContext(ctx, query, startDate, endDate, roomId)
	err := rows.Scan(&numOfRows)

	if err != nil {
		return false, err
	}
	if numOfRows > 0 {
		return false, nil
	} else {
		return true, nil
	}
}
