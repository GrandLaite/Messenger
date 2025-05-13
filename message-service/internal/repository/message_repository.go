package repository

import "database/sql"

type Message struct {
	ID                int
	SenderNickname    string
	RecipientNickname string
	Content           string
	CreatedAt         string
}

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(d *sql.DB) *MessageRepository {
	return &MessageRepository{db: d}
}

func (r *MessageRepository) Create(sender, recipient, content string) (Message, error) {
	q := `INSERT INTO messages (sender_nickname, recipient_nickname, content)
	      VALUES ($1, $2, $3)
	      RETURNING id, sender_nickname, recipient_nickname, content, created_at`
	var m Message
	err := r.db.QueryRow(q, sender, recipient, content).
		Scan(&m.ID, &m.SenderNickname, &m.RecipientNickname, &m.Content, &m.CreatedAt)
	return m, err
}

func (r *MessageRepository) GetByID(id int) (Message, error) {
	q := `SELECT id, sender_nickname, recipient_nickname, content, created_at
	      FROM messages WHERE id = $1`
	var m Message
	err := r.db.QueryRow(q, id).
		Scan(&m.ID, &m.SenderNickname, &m.RecipientNickname, &m.Content, &m.CreatedAt)
	return m, err
}

func (r *MessageRepository) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM messages WHERE id = $1`, id)
	return err
}

func (r *MessageRepository) GetConversation(u1, u2 string) ([]Message, error) {
	q := `SELECT id, sender_nickname, recipient_nickname, content, created_at
	      FROM messages
	      WHERE (sender_nickname = $1 AND recipient_nickname = $2)
	         OR (sender_nickname = $2 AND recipient_nickname = $1)
	      ORDER BY created_at`
	rows, err := r.db.Query(q, u1, u2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Message
	for rows.Next() {
		var m Message
		if err = rows.Scan(&m.ID, &m.SenderNickname, &m.RecipientNickname, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	return res, nil
}

func (r *MessageRepository) GetDialogs(nickname string) ([]string, error) {
	q := `SELECT DISTINCT partner FROM (
	        SELECT recipient_nickname AS partner FROM messages WHERE sender_nickname = $1
	        UNION
	        SELECT sender_nickname   AS partner FROM messages WHERE recipient_nickname = $1
	      ) AS tmp`
	rows, err := r.db.Query(q, nickname)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var partners []string
	for rows.Next() {
		var p string
		if err = rows.Scan(&p); err != nil {
			return nil, err
		}
		partners = append(partners, p)
	}
	return partners, nil
}
