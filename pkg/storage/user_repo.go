package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           string
	Email        string
	PasswordHash []byte
	DisplayName  sql.NullString
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserRepository struct {
	db         *sql.DB
	isSQLite   bool
}

func NewUserRepository(db *sql.DB) *UserRepository {
	// Detect database type by checking driver name
	isSQLite := false
	if db != nil {
		driver := db.Driver()
		if driverWithName, ok := driver.(interface{ DriverName() string }); ok {
			driverName := driverWithName.DriverName()
			isSQLite = driverName == "sqlite" || driverName == "sqlite3"
		}
	}
	
	return &UserRepository{
		db:       db,
		isSQLite: isSQLite,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, u *User) error {
	// Always generate ID first to ensure it's never empty
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	
	if r.isSQLite {
		// SQLite doesn't support RETURNING, so use explicit ID
		query := `
		INSERT INTO users (id, email, password_hash, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
		
		_, err := r.db.ExecContext(ctx, query, u.ID, u.Email, u.PasswordHash, u.DisplayName)
		if err != nil {
			return err
		}
		
		// Fetch created_at and updated_at
		query = `SELECT created_at, updated_at FROM users WHERE id = $1`
		err = r.db.QueryRowContext(ctx, query, u.ID).Scan(&u.CreatedAt, &u.UpdatedAt)
		return err
	}
	
	// PostgreSQL: use explicit ID with RETURNING clause
	query := `
	INSERT INTO users (id, email, password_hash, display_name)
	VALUES ($1, $2, $3, $4)
	RETURNING created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, u.ID, u.Email, u.PasswordHash, u.DisplayName).
		Scan(&u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	const query = `
	SELECT id, email, password_hash, display_name, created_at, updated_at
	FROM users WHERE email = $1`
	u := User{}
	var id sql.NullString
	err := r.db.QueryRowContext(ctx, query, email).
		Scan(&id, &u.Email, &u.PasswordHash, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	
	// Handle NULL ID (for users created before fix)
	if !id.Valid || id.String == "" {
		// Generate ID and update the record
		u.ID = uuid.New().String()
		updateQuery := `UPDATE users SET id = $1 WHERE email = $2`
		_, updateErr := r.db.ExecContext(ctx, updateQuery, u.ID, email)
		if updateErr != nil {
			// If update fails, still return user with generated ID
			// (the ID won't be saved but at least login won't crash)
		}
	} else {
		u.ID = id.String
	}
	
	return &u, nil
}
