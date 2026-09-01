package database

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// Example showing usage of sqlmock — adapt to your DB wrapper types.
func TestGetTenantByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	// Replace with your query and expected columns.
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "tenant-a")

	mock.ExpectQuery("^SELECT id, name FROM tenants WHERE id = \\\\$1$").
		WithArgs(1).
		WillReturnRows(rows)

	// Wrap db with your repository adapter, for example NewRepo(sqlx.NewDb(db, "postgres")).
	// repo := NewRepo(db)
	// tenant, err := repo.GetTenantByID(1)
	// if err != nil { t.Fatalf("unexpected error: %v", err) }
	// if tenant.Name != "tenant-a" { t.Fatalf("unexpected tenant name: %v", tenant.Name) }

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %v", err)
	}
}
