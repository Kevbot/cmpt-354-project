package postgresql

import (
	"database/sql"
	"draco/models"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type PlayerModel struct {
	DB *sqlx.DB
}

func (m *PlayerModel) Insert(username, password, name string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO Player (username, password, name)
	VALUES($1, $2, $3)`

	_, err = m.DB.Exec(stmt, username, string(hashedPassword), name)
	if err != nil {
		var postgresError *pq.Error
		if errors.As(err, &postgresError) {
			if postgresError.Code.Name() == "unique_violation" {
				if strings.Contains(postgresError.Message, "player_pkey") {
					return models.ErrDuplicateUsername
				}
			}
		}
		return err
	}

	return nil
}

// Authenticate verifies that `password` matches the password stored for
// the player identified by `username`, should they exist.
func (m *PlayerModel) Authenticate(username, password string) (string, error) {
	var storedUsername string
	var hashedPassword []byte

	stmt := "SELECT username, password FROM Player WHERE username = $1"
	row := m.DB.QueryRow(stmt, username)
	if err := row.Scan(&storedUsername, &hashedPassword); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", models.ErrNoRecord
		} else {
			return "", err
		}
	}

	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", bcrypt.ErrMismatchedHashAndPassword
		} else {
			return "", err
		}
	}

	return storedUsername, nil
}

// Get attempts to retrieve a player entity from the database with a
// username equal to `username`. Does not return the player's password.
func (m *PlayerModel) Get(username string) (*models.Player, error) {
	var storedUsername string
	var storedName string

	stmt := "SELECT username, name FROM Player WHERE username = $1"
	row := m.DB.QueryRow(stmt, username)
	if err := row.Scan(&storedUsername, &storedName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.Player{}, models.ErrNoRecord
		} else {
			return &models.Player{}, err
		}
	}

	p := &models.Player{
		Username: storedUsername,
		Name:     storedName,
	}

	return p, nil
}

// UpdatePassword attempts to update the password belonging to the
// Player identified by `username`.
func (m *PlayerModel) UpdatePassword(username, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	stmt := "UPDATE Player SET password = $2 WHERE username = $1"
	res, err := m.DB.Exec(stmt, username, hashedPassword)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return models.ErrUpdateSingleRecord
	}

	return nil
}

// Delete attempts to delete a player identified by `username`.
func (m *PlayerModel) Delete(username string) error {
	stmt := "DELETE FROM Player WHERE username = $1"

	res, err := m.DB.Exec(stmt, username)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return models.ErrDeleteSingleRecord
	}

	return nil
}
