package main

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type PostgreConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func (cfg PostgreConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)
}

func main() {
	cfg := PostgreConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "baloo",
		Password: "junglebook",
		Database: "lenslocked",
		SSLMode:  "disable",
	}

	db, err := sql.Open("pgx", cfg.String())
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected!")

	// Create table ...
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS users (
      id SERIAL PRIMARY KEY,
      name TEXT,
      email TEXT UNIQUE NOT NULL
    );

    CREATE TABLE IF NOT EXISTS orders (
      id SERIAL PRIMARY KEY,
      user_id INT NOT NULL,
      amount INT,
      description TEXT
    )
  `)
	if err != nil {
		panic(err)
	}
	fmt.Println("Tables created.")

	//##############################################################
	// Insert some data ...
	// notice that $1, $2 act as placeholders
	//name := "Jon Calhoun"
	//email := "jon@calhoun.io"

	//_, err = db.Exec(`
	//  INSERT INTO users (name, email)
	//  VALUES ($1, $2);`, name, email)
	//if err != nil {
	//	panic(err)
	//}

	//fmt.Println("User created.")

	//###############################################################
	// Insert data and get the id back.
	// This cannot be done with Exec but we can use QueryRow
	//name := "Mike Bisbin"
	//email := "mike@bisbin.com"

	//row := db.QueryRow(`
	//  INSERT INTO users (name, email)
	//  VALUES ($1, $2) RETURNING id;`, name, email)

	//var id int
	//// Remember in Go, function arguments are always passed by value
	//// (a copy of the original). So if you want to modify a variable,
	//// you must pass a pointer to that variable.
	//err = row.Scan(&id)

	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("User created. id = ", id)

	//##############################################################

	// Query data with QueryRow
	id := 2
	row := db.QueryRow(`
    SELECT name, email
    FROM users
    WHERE id=$1;`, id)

	var name, email string
	err = row.Scan(&name, &email)
	if err == sql.ErrNoRows {
		fmt.Println("Error, no rows")
	}
	if err != nil {
		panic(err)
	}
	fmt.Printf("User information: name=%s, email=%s\n", name, email)
}
