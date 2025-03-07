package main

import (
	"fmt"

	"github.com/etaseq/lenslocked/models"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func (cfg PostgresConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)
}

func main() {
	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected!")

	us := models.UserService{
		DB: db,
	}
	user, err := us.Create("bob4@bob.com", "bob123")
	if err != nil {
		panic(err)
	}
	fmt.Println(user)

	// Create table ...
	//_, err = db.Exec(`
	//  CREATE TABLE IF NOT EXISTS users (
	//    id SERIAL PRIMARY KEY,
	//    name TEXT,
	//    email TEXT UNIQUE NOT NULL
	//  );

	//  CREATE TABLE IF NOT EXISTS orders (
	//    id SERIAL PRIMARY KEY,
	//    user_id INT NOT NULL,
	//    amount INT,
	//    description TEXT
	//  )
	//`)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Tables created.")

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
	//id := 2
	//row := db.QueryRow(`
	//  SELECT name, email
	//  FROM users
	//  WHERE id=$1;`, id)

	//var name, email string
	//err = row.Scan(&name, &email)
	//if err == sql.ErrNoRows {
	//	fmt.Println("Error, no rows")
	//}
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("User information: name=%s, email=%s\n", name, email)

	//#################################################################

	// Add some fake orders
	//userID := 1

	//for i := 1; i <= 5; i++ {
	//	amount := i * 100
	//	desc := fmt.Sprintf("Fake order #%d", i)
	//	_, err := db.Exec(`
	//    INSERT INTO orders(user_id, amount, description)
	//    VALUES($1, $2, $3)`, userID, amount, desc)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//fmt.Println("Created fake orders.")

	//##################################################################

	// Query multiple records
	//type Order struct {
	//	ID          int
	//	UserID      int
	//	Amount      int
	//	Description string
	//}

	//var orders []Order
	//userID := 1

	//rows, err := db.Query(`
	//  SELECT id, amount, description
	//  FROM orders
	//  WHERE user_id=$1`, userID)
	//if err != nil {
	//	panic(err)
	//}
	//defer rows.Close()

	//// Loop through the record.
	//// Next() was designed by the designers of the sql package
	//// to start from the first entry although you would expect
	//// to start from the second since it is Next.
	//// So it includes all the records in the query.
	//for rows.Next() {
	//	var order Order
	//	order.UserID = userID
	//	err := rows.Scan(&order.ID, &order.Amount, &order.Description)

	//	if err != nil {
	//		panic(err)
	//	}
	//	orders = append(orders, order)
	//}

	//err = rows.Err()
	//if err != nil {
	//	panic(err)
	//}

	//fmt.Println("Orders: ", orders)
}
