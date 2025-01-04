package dbutils_test

import (
	"errors"
	"testing"

	"gurch101.github.io/go-web/pkg/dbutils"
)

func TestGetByID(t *testing.T) {
	t.Parallel()
	db := dbutils.SetupTestDB(t)

	defer func() {
		closeErr := db.Close()
		if closeErr != nil {
			t.Fatalf("Failed to close database connection: %v", closeErr)
		}
	}()

	t.Run("successful retrieval", func(t *testing.T) {
		var name, email string
		fields := map[string]any{
			"user_name": &name,
			"email":     &email,
		}

		err := dbutils.GetByID(db, "users", 1, fields)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if name != "admin" {
			t.Errorf("Expected name 'admin', got '%s'", name)
		}

		if email != "admin@acme.com" {
			t.Errorf("Expected email 'admin@acme.com', got '%s'", email)
		}
	})

	t.Run("negative ID", func(t *testing.T) {
		var name string
		fields := map[string]any{
			"user_name": &name,
		}

		err := dbutils.GetByID(db, "users", -1, fields)
		if !errors.Is(err, dbutils.ErrRecordNotFound) {
			t.Errorf("Expected ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("non-existent record", func(t *testing.T) {
		var name string
		fields := map[string]any{
			"user_name": &name,
		}

		err := dbutils.GetByID(db, "users", 999, fields)
		if !errors.Is(err, dbutils.ErrRecordNotFound) {
			t.Errorf("Expected ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("empty fields map", func(t *testing.T) {
		fields := map[string]any{}

		err := dbutils.GetByID(db, "users", 1, fields)
		if err == nil {
			t.Error("Expected error for empty fields map, got nil")
		}
	})
}
