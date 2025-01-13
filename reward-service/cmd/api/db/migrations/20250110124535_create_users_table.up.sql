CREATE TABLE IF NOT EXISTS users(
                       id serial PRIMARY KEY,
                       email VARCHAR(255) UNIQUE NOT NULL,
                       first_name VARCHAR(100),
                       last_name VARCHAR(100),
                       password VARCHAR(255) NOT NULL,
                       active INT NOT NULL DEFAULT 1,
                       score INT NOT NULL DEFAULT 0,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
