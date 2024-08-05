package serveur

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gorilla/websocket"
)

const DataSourceName = "../backend/serveur/db/data.sql"

type DataBase struct {
	DB *sql.DB
}

type User struct {
	Id   string
	Name string
}

func (db *DataBase) SaveMessage(Db *sql.DB, online map[string]*websocket.Conn, SenderID, ReceiverID, content string) error {
	db.DB = db.OpenDB(DataSourceName)
	defer db.DB.Close()
	inConversation := make(map[string]*websocket.Conn)
	for id, wsconnOfUser := range online {
		if (id == SenderID) || (id == ReceiverID) {
			inConversation[id] = wsconnOfUser
		}
	}

	// Insert the message into the database and get the timestamp of the inserted message
	result, err := db.DB.Exec("INSERT INTO chat_message (sender_user_id, receiver_user_id, message) VALUES (?, ?, ?)", SenderID, ReceiverID, content)
	if err != nil {
		log.Fatal(err)
		return err
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
		return err
	}

	var messageTime string
	err = db.DB.QueryRow("SELECT date FROM chat_message WHERE id = ?", messageID).Scan(&messageTime)
	if err != nil {
		log.Fatal(err)
		return err
	}

	var usernameSender, usernameReceiver string

	// Get the sender's username
	err = db.DB.QueryRow("SELECT username FROM users WHERE id = ?", SenderID).Scan(&usernameSender)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// Get the receiver's username
	err = db.DB.QueryRow("SELECT username FROM users WHERE id = ?", ReceiverID).Scan(&usernameReceiver)
	if err != nil {
		log.Fatal(err)
		return err
	}

	msg := MessageInfo{
		SenderID:     SenderID,
		ReceiverID:   ReceiverID,
		Content:      content,
		SenderName:   usernameSender,
		ReceiverName: usernameReceiver,
		SentAt:       messageTime,
	}
	fmt.Println(msg)
	response := Response{
		DataName: "newmessage",
		Data:     msg,
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Println("Error converting response to JSON:", err)
		return err
	}

	// Broadcast messages
	for _, wsconnOfUserInConv := range inConversation {
		if err := wsconnOfUserInConv.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Println("Broadcast error sending data via WebSocket:", err)
			return err
		}
	}
	return nil
}

func (db *DataBase) GetMessages(conn *websocket.Conn, senderID string, receiverID string) ([]MessageInfo, error) {
	var messages []MessageInfo
	db.DB = db.OpenDB(DataSourceName)
	defer db.DB.Close()

	// Get the sender and receiver usernames
	var senderName, receiverName string
	err := db.DB.QueryRow("SELECT username FROM users WHERE id = ?", senderID).Scan(&senderName)
	if err != nil {
		return nil, err
	}

	err = db.DB.QueryRow("SELECT username FROM users WHERE id = ?", receiverID).Scan(&receiverName)
	if err != nil {
		return nil, err
	}

	// Fetch messages between sender and receiver
	rows, err := db.DB.Query(
		"SELECT sender_user_id, receiver_user_id, message, date FROM chat_message WHERE (sender_user_id = ? AND receiver_user_id = ?) OR (sender_user_id= ? AND receiver_user_id = ?)",
		senderID, receiverID, receiverID, senderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var msg MessageInfo
		err := rows.Scan(&msg.SenderID, &msg.ReceiverID, &msg.Content, &msg.SentAt)
		if err != nil {
			return nil, err
		}
		// Assign the usernames to the message
		if msg.SenderID == senderID {
			msg.SenderName = senderName
			msg.ReceiverName = receiverName
		} else {
			msg.SenderName = receiverName
			msg.ReceiverName = senderName
		}
		messages = append(messages, msg)
	}
	fmt.Println(messages)

	response := Response{
		DataName: "conversation",
		Data:     messages,
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Println("Error converting response to JSON:", err)
		return nil, err
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Println("Error sending data via WebSocket:", err)
		return nil, err
	}
	return messages, nil
}
func (db *DataBase) GetAllUsers(conn *websocket.Conn, id string) error {
	db.DB = db.OpenDB(DataSourceName)
	defer db.DB.Close()

	query := `
        SELECT id, username
        FROM users
        WHERE id != ?
        ORDER BY username ASC
    `
	rows, err := db.DB.Query(query, id)
	if err != nil {
		log.Println("Erreur lors de l'exécution de la requête SQL:", err)
		return err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Id, &user.Name); err != nil {
			log.Println("Erreur lors de la lecture des données utilisateur:", err)
			return err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		log.Println("Erreur lors de l'itération des lignes utilisateur:", err)
		return err
	}

	response := Response{
		DataName: "AllUsers",
		Data:     users,
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Println("Erreur lors de la conversion des utilisateurs en JSON:", err)
		return err
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Println("Erreur lors de l'envoi des données utilisateur via WebSocket:", err)
		return err
	}

	return nil
}

func (db *DataBase) GetAllPub(conn *websocket.Conn) error {
	db.DB = db.OpenDB(DataSourceName)
	defer db.DB.Close()

	query := `
        SELECT p.id, u.username, p.categories, p.title, p.content, p.creation_date
        FROM post_forum p
        JOIN users u ON p.user_id = u.id
    `
	rows, err := db.DB.Query(query)
	if err != nil {
		log.Println("Erreur lors de la récupération des publications depuis la base de données:", err)
		return err
	}
	defer rows.Close()

	var pubs []Pub

	for rows.Next() {
		var publication Pub
		err := rows.Scan(&publication.Id, &publication.UserNickname, &publication.Categories, &publication.Title, &publication.Content, &publication.CreationDate)
		if err != nil {
			log.Println("Erreur lors de la lecture des données de publication:", err)
			continue
		}

		commentQuery := `
            SELECT u.username, c.content, c.creation_date
            FROM forum_comment c
            JOIN users u ON c.user_id = u.id
            WHERE c.post_id = ?
        `
		commentRows, err := db.DB.Query(commentQuery, publication.Id)
		if err != nil {
			log.Println("Erreur lors de la récupération des commentaires depuis la base de données:", err)
			continue
		}
		defer commentRows.Close()

		for commentRows.Next() {
			var comment Comments
			err := commentRows.Scan(&comment.Username, &comment.Content, &comment.CreationDate)
			if err != nil {
				log.Println("Erreur lors de la lecture des données de commentaire:", err)
				continue
			}

			publication.Commentaire = append(publication.Commentaire, comment)
		}

		pubs = append(pubs, publication)
	}

	if err := rows.Err(); err != nil {
		log.Println("Erreur lors de l'itération des lignes de publication:", err)
		return err
	}

	response := Response{
		DataName: "publications",
		Data:     pubs,
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Println("Erreur lors de la conversion des publications en JSON:", err)
		return err
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Println("Erreur lors de l'envoi des données des publications via WebSocket:", err)
		return err
	}
	return nil
}
func (db *DataBase) InsertNewComment(UserId, PubId, Comment string) error {
	db.DB = db.OpenDB(DataSourceName)
	defer db.DB.Close()
	log.Println("Inserting comment into database:", UserId, PubId, Comment) // Add log here

	query := `
    INSERT INTO forum_comment (user_id, post_id, content) VALUES (?, ?, ?)
`
	_, err := db.DB.Exec(query, UserId, PubId, Comment)
	if err != nil {
		log.Println("Error inserting comment:", err)
		return err
	}
	return nil
}

func (db *DataBase) InsertPost(userid string, categorie string, title string, content string) error {
	var err error
	db.DB = db.OpenDB(DataSourceName)
	defer db.DB.Close()

	stmt := `INSERT INTO post_forum (user_id, categories, title, content) VALUES (?, ?, ?, ?)`
	_, err = db.DB.Exec(stmt, userid, categorie, title, content)
	if err != nil {
		return err
	}

	return nil
}

func (db *DataBase) SetTokenToUser(id int, cookiesToken string) error {
	var err error

	db.DB = db.OpenDB(DataSourceName)

	defer db.DB.Close()
	stmt := `UPDATE users SET token = ? WHERE id = ?`

	_, err = db.DB.Exec(stmt, cookiesToken, id)
	if err != nil {
		return err
	}
	return nil
}

func (db *DataBase) InsertUser(username, firstName, lastName, gender, mail, password string) error {
	db.DB = db.OpenDB(DataSourceName)
	defer db.DB.Close()

	// Check if username exist
	var count int = 0
	err := db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("le nom d'utilisateur existe déjà")
	}

	// Check if mail exist
	count = 0
	err = db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE mail = ?", mail).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("l'adresse e-mail existe déjà")
	}

	// finaly insert user to db whit his info
	stmt := `INSERT INTO users (username, first_name, last_name, gender, mail, password) VALUES (?, ?, ?, ?, ?, ?)`
	_, err = db.DB.Exec(stmt, username, firstName, lastName, gender, mail, password)
	if err != nil {
		return err
	}
	return nil
}

func (db *DataBase) CheckUser(username, password string) (int, error) {
	db.DB = db.OpenDB(DataSourceName)
	defer db.DB.Close()

	// select the correct statment
	var stmt, mailOrNick string
	if strings.Contains(username, "@") {
		mailOrNick = "mail"
		stmt = `SELECT id FROM users WHERE mail = ? AND password = ?`
	} else {
		mailOrNick = "nickname"
		stmt = `SELECT id FROM users WHERE username = ? AND password = ?`
	}

	// check if user existe in db
	var id int
	// this statment work also stmt = `SELECT id FROM users WHERE`+mailOrNick+`= ? AND password = ?` (update the if stmt)
	row := db.DB.QueryRow(stmt, username, password)
	err := row.Scan(&id)
	if err != nil && err.Error() == "sql: no rows in result set" {
		return 0, errors.New(mailOrNick + " or password invalid")
	} else if err != nil {
		return 0, err
	}
	return id, nil
}

func (db *DataBase) InitDB() error {
	db.DB = db.OpenDB(DataSourceName)
	// if error then os.exit1
	defer db.DB.Close()
	UsersTable := `CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT UNIQUE NOT NULL,
        first_name TEXT NOT NULL,
        last_name TEXT NOT NULL,
        gender TEXT NOT NULL,
        mail TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL,
        creation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		token TEXT NULL
    );`
	_, err := db.DB.Exec(UsersTable)
	if err != nil {
		return err
	}
	PostTable := `CREATE TABLE IF NOT EXISTS post_forum (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        categories TEXT NOT NULL,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
        creation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users (id)
    );`
	_, err = db.DB.Exec(PostTable)
	if err != nil {
		return err
	}
	CommentTable := `CREATE TABLE IF NOT EXISTS forum_comment (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        post_id INTEGER NOT NULL,
        content TEXT NOT NULL,
        creation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users (id),
        FOREIGN KEY (post_id) REFERENCES post_forum (id)
    );`
	_, err = db.DB.Exec(CommentTable)
	if err != nil {
		return err
	}

	ChatTable := `CREATE TABLE IF NOT EXISTS chat_message (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        sender_user_id INTEGER NOT NULL,
		receiver_user_id INTEGER NOT NULL,
		message TEXT NOT NULL,
        date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (sender_user_id) REFERENCES users (id),
        FOREIGN KEY (receiver_user_id) REFERENCES users (id)
    );`
	_, err = db.DB.Exec(ChatTable)
	if err != nil {
		return err
	}
	return nil
}

func (db *DataBase) OpenDB(dsn string) *sql.DB {
	var err error
	db.DB, err = sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return db.DB
}

// func (db *DataBase) GetMessages(conn *websocket.Conn, senderID string, receiverID string) ([]MessageInfo, error) {
// 	db.DB = db.OpenDB(DataSourceName)
// 	defer db.DB.Close()
// 	rows, err := db.DB.Query("SELECT sender_user_id, receiver_user_id, message FROM chat_message WHERE (sender_user_id = ? AND receiver_user_id = ?) OR (sender_user_id= ? AND receiver_user_id = ?)", senderID, receiverID, receiverID, senderID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	var messages []MessageInfo
// 	for rows.Next() {
// 		var msg MessageInfo
// 		err := rows.Scan(&msg.SenderID, &msg.ReceiverID, &msg.Content)
// 		if err != nil {
// 			return nil, err
// 		}
// 		messages = append(messages, msg)
// 	}
// 	response := Response{
// 		DataName: "conversation",
// 		Data:     messages,
// 	}
// 	data, err := json.Marshal(response)
// 	if err != nil {
// 		log.Println("Erreur lors de la conversion des utilisateurs en JSON:", err)
// 		return nil, err
// 	}
// 	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
// 		log.Println("Erreur lors de l'envoi des données des utilisateurs via WebSocket:", err)
// 		return nil, err
// 	}
// 	return messages, nil
// }
