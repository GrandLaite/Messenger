package repository

import "database/sql"

type User struct {
	ID           int
	Username     string
	PasswordHash string
	Role         string
	Email        string
	Nickname     string
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(d *sql.DB) *UserRepository {
	return &UserRepository{db: d}
}

func (r *UserRepository) Create(u User) (int, error) {
	q := `INSERT INTO users (username, password_hash, role, email, nickname)
	      VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var id int
	err := r.db.QueryRow(q, u.Username, u.PasswordHash, u.Role, u.Email, u.Nickname).Scan(&id)
	return id, err
}

func (r *UserRepository) GetByUsername(username string) (User, error) {
	var u User
	q := `SELECT id, username, password_hash, role, email, nickname
	      FROM users WHERE username = $1`
	err := r.db.QueryRow(q, username).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.Email, &u.Nickname)
	return u, err
}

func (r *UserRepository) GetByNickname(nickname string) (User, error) {
	var u User
	q := `SELECT id, username, password_hash, role, email, nickname
	      FROM users WHERE nickname = $1`
	err := r.db.QueryRow(q, nickname).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.Email, &u.Nickname)
	return u, err
}
