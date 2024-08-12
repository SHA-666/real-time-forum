package main

import (
	"log"
	"net/http"

	serv "real/serveur"

	_ "github.com/mattn/go-sqlite3"
)

var db serv.DataBase

func main() {
	
	err := db.InitDB()
	if err != nil {
		log.Println("Error while opening db err = :", err)
		return
	}
	defer db.DB.Close()

	//serve static files
	http.Handle("/front/", http.StripPrefix("/front/", http.FileServer(http.Dir("../front/"))))

	http.HandleFunc("/", serv.HomeHandler)
	http.HandleFunc("/ws", serv.HandleWS)
	log.Println("WebSocket server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
