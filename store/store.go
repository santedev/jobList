package store

import (
	"database/sql"
	"fmt"
	"jobList/config"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

var DB *Storage

func Sqlconfig() string {
	return fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=%s",
		config.Envs.DBUser,
		config.Envs.DBPassword,
		config.Envs.DBHost,
		config.Envs.DBName,
		config.Envs.DBsslMode)
}

func NewStore(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) GetSavedJobOffers(uid string) ([]string, error) {
	rows, err := s.db.Query("SELECT offer_link from offers WHERE uid = $1", uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []string
	for rows.Next() {
		var offerLink string
		if err := rows.Scan(&offerLink); err != nil {
			return nil, err
		}
		offers = append(offers, offerLink)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return offers, nil
}

func (s *Storage) UnsaveJobOffer(uid string, jobLink string) error {
	result, err := s.db.Exec("DELETE FROM offers WHERE uid = $1 AND offer_link = $2", uid, jobLink)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows deleted; offer may not exist")
	}
	return nil
}

func (s *Storage) SaveJobOffer(uid string, jobLink string) error {
	exits, err := s.JobOfferExists(uid, jobLink)
	if err != nil {
		return err
	}
	if exits {
		return fmt.Errorf("tried to save an offer that is already saved")
	}
	result, err := s.db.Exec("INSERT INTO offers (uid, offer_link) VALUES($1, $2)", uid, jobLink)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected; offer is not saved")
	}
	return nil
}

func (s *Storage) JobOfferExists(uid string, jobLink string) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS (SELECT 1 FROM offers WHERE uid = $1 AND offer_link = $2)", uid, jobLink).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
