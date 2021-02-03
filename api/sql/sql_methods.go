package main

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type UserAPI struct {
    Email   string    `json:"email"`
    APIKey string `json:"apiKey"`
}

func main() {
	fmt.Println("Go MySQL Tutorial")

    // Open up our database connection.
    // I've set up a database on my local machine using phpmyadmin.
    // The database is called testDb
    db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/trackerapi")

    // if there is an error opening the connection, handle it
    if err != nil {
        panic(err.Error())
    }


	// Execute the query
	results, err := db.Query("SELECT * FROM UserAPI")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	for results.Next() {
		var userapi UserAPI
		// for each row, scan the result into our tag composite object
		err = results.Scan(&userapi.Email, &userapi.APIKey)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
				// and then print out the tag's Name attribute
		log.Printf(userapi.Email)
		log.Printf(userapi.APIKey)
	}

    // defer the close till after the main function has finished
    // executing
    defer db.Close()
}