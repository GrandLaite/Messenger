CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    nickname VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    sender_nickname VARCHAR(50) NOT NULL,
    recipient_nickname VARCHAR(50) NOT NULL,
    content TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_sender
      FOREIGN KEY (sender_nickname) REFERENCES users (nickname) ON DELETE CASCADE,
    CONSTRAINT fk_recipient
      FOREIGN KEY (recipient_nickname) REFERENCES users (nickname) ON DELETE CASCADE
);

