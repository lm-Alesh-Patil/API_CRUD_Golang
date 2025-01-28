package main

import (
	"database/sql"
	"log"
	"net/http"

	"web_request_methods/server"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/userdb?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	srv := server.NewServer(db)
	router := srv.InitRouter()

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
