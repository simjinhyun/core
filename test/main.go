package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:Tldrmf#2013@tcp(10.0.0.200:3306)/testdb?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		panic(err)
	}

	rows, err := db.Query("SELECT id, name, age, phone, address, created_at FROM address_book;")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id        int
			name      string
			age       int
			phone     string
			address   string
			createdAt time.Time
		)
		if err := rows.Scan(&id, &name, &age, &phone, &address, &createdAt); err != nil {
			panic(err)
		}
		fmt.Println("UTC :", createdAt.UTC().Format("2006-01-02 15:04:05"))
		fmt.Println("KST :", createdAt.In(time.FixedZone("KST", 9*60*60)).Format("2006-01-02 15:04:05"))

	}

	if err := rows.Err(); err != nil {
		panic(err)
	}
}
