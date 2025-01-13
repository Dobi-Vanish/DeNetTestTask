package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"reward-service/data"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "82"

var counts int64

type Config struct {
	Repo   data.Repository
	Client *http.Client
}

// main starts the server and establishing connection to database
func main() {
	log.Println("Starting reward service")

	// connect to DB
	conn := connectToDB()
	if conn == nil {
		log.Panic("Can't connect to Postgres!")
	}

	// set up config
	app := Config{
		Client: &http.Client{},
	}
	app.setupRepo(conn)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

// // openDB establishes a connection to the PostgreSQL database using the provided Data Source Name (DSN)
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx/v4", dsn)
	fmt.Println(db)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// connectToDB connect to Postgres with provided dsn
func connectToDB() *sql.DB {
	dsn := "host=postgres port=5432 dbname=users user=postgres password=password"

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready ...")
			counts++
		} else {
			log.Println("Connected to Postgres!")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two seconds....")
		time.Sleep(2 * time.Second)
		continue
	}
}

// setupRepo sets new postgres repository
func (app *Config) setupRepo(conn *sql.DB) {
	db := data.NewPostgresRepository(conn)
	app.Repo = db
}
