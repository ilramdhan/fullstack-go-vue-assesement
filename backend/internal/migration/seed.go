package migration

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Seed(db *sql.DB) error {
	if err := seedUsers(db); err != nil {
		return fmt.Errorf("seed users: %w", err)
	}
	if err := seedPayments(db); err != nil {
		return fmt.Errorf("seed payments: %w", err)
	}
	return nil
}

func seedUsers(db *sql.DB) error {
	var n int
	if err := db.QueryRow(`SELECT COUNT(1) FROM users`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	rows := []struct{ email, role string }{
		{"cs@test.com", "cs"},
		{"operation@test.com", "operation"},
	}
	for _, u := range rows {
		if _, err := db.Exec(
			`INSERT INTO users(email, password_hash, role) VALUES (?, ?, ?)`,
			u.email, string(hash), u.role,
		); err != nil {
			return err
		}
	}
	return nil
}

var merchants = []string{
	"Tokopedia", "Shopee", "Bukalapak", "Blibli", "Lazada",
	"GrabFood", "GoFood", "Traveloka", "Tiket.com", "Pegipegi",
	"Sociolla", "Zalora", "Sephora", "Watsons", "Guardian",
	"Indomaret", "Alfamart", "Hypermart", "Transmart", "Ranch Market",
}

func seedPayments(db *sql.DB) error {
	var n int
	if err := db.QueryRow(`SELECT COUNT(1) FROM payments`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	r := rand.New(rand.NewSource(42))
	statuses := buildStatusMix()

	stmt, err := db.Prepare(
		`INSERT INTO payments(id, merchant, amount, currency, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC()
	for _, status := range statuses {
		merchant := merchants[r.Intn(len(merchants))]
		amount := int64(10_000+r.Intn(4_990_000)) * 100
		offset := time.Duration(r.Int63n(int64(30 * 24 * time.Hour)))
		createdAt := now.Add(-offset).Format(time.RFC3339)

		if _, err := stmt.Exec(uuid.NewString(), merchant, amount, "IDR", status, createdAt); err != nil {
			return err
		}
	}
	return nil
}

func buildStatusMix() []string {
	mix := make([]string, 0, 50)
	for i := 0; i < 38; i++ {
		mix = append(mix, "completed")
	}
	for i := 0; i < 7; i++ {
		mix = append(mix, "processing")
	}
	for i := 0; i < 5; i++ {
		mix = append(mix, "failed")
	}
	r := rand.New(rand.NewSource(7))
	r.Shuffle(len(mix), func(i, j int) { mix[i], mix[j] = mix[j], mix[i] })
	return mix
}
