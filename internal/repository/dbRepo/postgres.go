package dbrepo

import (
	"context"
	"fmt"
	"github.com/leoashish99/bookings/internal/models"
	"golang.org/x/crypto/bcrypt"
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
func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomId(startDate, endDate time.Time, roomId int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select count(*)
              from room_restrictions
              where end_date > $1 and start_date < $2
              and room_id = $3`

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

// Returns a slice of
// available rooms if any is present
func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select r.id, r.room_name
              from rooms r
              where r.id not in(
                  select rr.room_id
                  from room_restrictions rr
                  where  $1< rr.end_date AND $2 > rr.start_date
              )
	`

	rows, err := m.DB.QueryContext(ctx, query, start, end)

	if err != nil {
		return []models.Room{}, err
	}
	var rooms []models.Room
	for rows.Next() {
		var room string
		var id int
		err := rows.Scan(&id, &room)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, models.Room{
			ID:       id,
			RoomName: room,
		})
	}
	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

func (m *postgresDBRepo) GetRoomById(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select r.id, r.room_name, r.created_at, r.updated_at
              from rooms r
              where r.id = $1
	`

	row := m.DB.QueryRowContext(ctx, query, id)

	if row.Err() != nil {
		return models.Room{}, row.Err()
	}
	var room string
	var createdAt time.Time
	var updatedAt time.Time

	err := row.Scan(&id, &room, &createdAt, &updatedAt)
	if err != nil {
		return models.Room{}, err
	}
	selectedRoom := models.Room{
		ID:        id,
		RoomName:  room,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return selectedRoom, nil
}

type User struct {
	ID          int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	AccessLevel int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (m *postgresDBRepo) GetUserById(ID int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select u.id, u.first_name, u.last_name, u.email , u.password, u.acess_level, u.created_at, u.updated_at
              from users u
              where u.id = $1
	`
	row := m.DB.QueryRowContext(ctx, query, ID)

	user := models.User{}

	err := row.Scan(user.ID, user.FirstName, user.LastName, user.Email, user.Password, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return user, err
	}
	return user, nil
}

func (m *postgresDBRepo) UpdateAUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update user 
              set first_name = $1 , last_name= $2, email = $3,
                  access_level = $4, updated_at = $5
    `
	_, err := m.DB.ExecContext(ctx, query, user.FirstName, user.LastName, user.Email, user.AccessLevel, time.Now())

	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id,password from users
				where email = $1`

	row := m.DB.QueryRowContext(ctx, query, email)

	var id int
	var hashedPassword string

	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return 0, "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))

	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", err
	} else {
		return id, testPassword, nil
	}

}
func (m *postgresDBRepo) GetAllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	select r.id, r.first_name, r.last_name, r.phone, r.email, r.start_date, r.end_date,
	       r.room_id, r.created_at, r.updated_at, r.processed, rm.id, rm.room_name
	from reservations r
	left join rooms rm
	on r.room_id = rm.id
	order by r.start_date asc
	`

	rows, err := m.DB.QueryContext(ctx, query)

	reservations := []models.Reservation{}
	if err != nil {
		return reservations, err
	}

	for rows.Next() {
		res := models.Reservation{}
		err := rows.Scan(&res.ID, &res.FirstName, &res.LastName, &res.Phone, &res.Email,
			&res.StartDate, &res.EndDate, &res.RoomID,
			&res.CreatedAt, &res.UpdatedAt, &res.Processed, &res.Room.ID, &res.Room.RoomName)
		if err != nil {
			fmt.Println(err)
			return reservations, err
		}
		reservations = append(reservations, res)
	}

	return reservations, nil
}

func (m *postgresDBRepo) GetAllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	select r.id, r.first_name, r.last_name, r.phone, r.email, r.start_date, r.end_date,
	       r.room_id, r.created_at, r.updated_at,r.processed, rm.id, rm.room_name
	from reservations r
	left join rooms rm
	on r.room_id = rm.id
	where r.processed = 0
	order by r.start_date asc
	`

	rows, err := m.DB.QueryContext(ctx, query)

	reservations := []models.Reservation{}
	if err != nil {
		return reservations, err
	}

	for rows.Next() {
		res := models.Reservation{}
		err := rows.Scan(&res.ID, &res.FirstName, &res.LastName, &res.Phone, &res.Email,
			&res.StartDate, &res.EndDate, &res.RoomID,
			&res.CreatedAt, &res.UpdatedAt, &res.Processed, &res.Room.ID, &res.Room.RoomName)
		if err != nil {
			fmt.Println(err)
			return reservations, err
		}
		reservations = append(reservations, res)
	}

	return reservations, nil
}

func (m *postgresDBRepo) GetReservation(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	select r.id, r.first_name, r.last_name, r.phone, r.email, r.start_date, r.end_date,
	       r.room_id, r.created_at, r.updated_at,r.processed, rm.id, rm.room_name
	from reservations r
	left join rooms rm
	on r.room_id = rm.id
	where r.id = $1
	`

	row := m.DB.QueryRowContext(ctx, query, id)

	reservation := models.Reservation{}

	err := row.Scan(&reservation.ID, &reservation.FirstName, &reservation.LastName, &reservation.Phone,
		&reservation.Email, &reservation.StartDate, &reservation.EndDate, &reservation.RoomID,
		&reservation.CreatedAt, &reservation.UpdatedAt, &reservation.Processed, &reservation.Room.ID,
		&reservation.Room.RoomName)

	if err != nil {
		fmt.Println(err)
		return reservation, err
	}
	return reservation, nil
}

func (m *postgresDBRepo) UpdateReservation(reservation models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update reservations
              set first_name = $1 , last_name= $2, email = $3,
                  phone = $4, updated_at = $5
			  where id = $6
    `
	_, err := m.DB.ExecContext(ctx, query, reservation.FirstName, reservation.LastName, reservation.Email, reservation.Phone, time.Now(), reservation.ID)

	if err != nil {
		return err
	}
	return nil
}
func (m *postgresDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `delete from reservations where id = $1`
	_, err := m.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) UpdateProcessedForReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	fmt.Println("processed = ", processed)
	query := `
			update reservations 
			set processed = $1
			where id = $2
`
	_, err := m.DB.ExecContext(ctx, query, processed, id)

	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
func (m *postgresDBRepo) AllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `select id, room_name, created_at, updated_at
               from rooms order by room_name`

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		return rooms, err
	}
	defer rows.Close()

	for rows.Next() {
		room := models.Room{}
		err := rows.Scan(&room.ID, &room.RoomName, &room.CreatedAt, &room.UpdatedAt)
		if err != nil {
			return rooms, nil
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}
