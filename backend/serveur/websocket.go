package serveur

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins
		},
	}
	userLock       sync.Mutex
	ConnectedUsers = make(map[string]*websocket.Conn)
	db             DataBase
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../front/index.html")
}

func addConnectedUser(userID string, conn *websocket.Conn) map[string]*websocket.Conn {
	userLock.Lock()
	ConnectedUsers[userID] = conn
	defer userLock.Unlock()
	return ConnectedUsers
}

func deleteConnectedUser(userID string) string {
	userLock.Lock()
	defer userLock.Unlock()
	delete(ConnectedUsers, userID)
	return userID
}

func HandleWS(w http.ResponseWriter, r *http.Request) {
	
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	var textErr, userID string
	conn.NetConn()
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		var data DataStruct
		if err := json.Unmarshal(msg, &data); err != nil {
			log.Println("Error decoding data:", err)
			continue
		}

		switch data.DataName {
		case "typing":
			// broadcast the typin
			broadcastTypinInProgress(ConnectedUsers, data.Typing.SenderID, data.Typing.ReceiverID, data.Typing.Leng)
		case "userconnection":
			// if  want redirect to login ( need to update showhome func jacascript)
			if data.UserConn.Id != "" {
				db.GetAllPub(conn)
				userID = data.UserConn.Id
				db.GetAllUsers(conn, userID)
				onlineUsr := addConnectedUser(userID, conn)
				broadcastOnlineUsers(onlineUsr, "online", "0")
			}
		case "register":
			err := registerRequest(&db, data.RegisterData.Nickname, data.RegisterData.FirstName, data.RegisterData.LastName, data.RegisterData.Gender, data.RegisterData.Email, data.RegisterData.Password)
			if err != nil {
				textErr = `{"success": false, "datas":"register" ,"error":"` + err.Error() + `"}`
				conn.WriteMessage(websocket.TextMessage, []byte(textErr))
			}
			conn.WriteMessage(websocket.TextMessage, []byte(`{"success": true , "datas":"register"}`))
		case "login":
			id, err := loginRequest(&db, data.LoginData.Nickname, data.LoginData.Password)
			if err != nil {
				textErr = `{"success": false,  "datas":"login", "error":"` + err.Error() + `"}`
				conn.WriteMessage(websocket.TextMessage, []byte(textErr))
				continue
			}
			newUUID := uuid.New()
			fmt.Println(newUUID.String())
			err = db.SetTokenToUser(id, newUUID.String())
			if err != nil {
				log.Fatal(err)
			}
			idToSend := strconv.Itoa(id)
			textErr := `{"success": true, "datas": "login", "id": "` + idToSend + `", "cookies": "` + newUUID.String() + `"}`
			conn.WriteMessage(websocket.TextMessage, []byte(textErr))
		case "creatPost":
			err := postRequest(&db, data.CreatPost.Id, data.CreatPost.Category, data.CreatPost.Title, data.CreatPost.Content)
			if err != nil {
				textErr = `{"success": false, "datas":"creatPost" ,"error":"` + err.Error() + `"}`
				conn.WriteMessage(websocket.TextMessage, []byte(textErr))
				continue
			}
			conn.WriteMessage(websocket.TextMessage, []byte(`{"success": true , "datas":"creatPost"}`))
			db.GetAllPub(conn)
		case "commentaire":
			log.Println("Received comment data:", data.NewComment) // Add log here
			err := newCommentRequest(&db, data.NewComment.User, data.NewComment.IDPub, data.NewComment.Comment)
			if err != nil {
				log.Println("Error inserting new comment:", err)
				textErr := `{"success": false, "datas":"newComment", "error":"` + err.Error() + `"}`
				conn.WriteMessage(websocket.TextMessage, []byte(textErr))
				continue
			}
			db.GetAllPub(conn)
		case "profil":
			fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$", data.IdOfUser.UserId)
		case "newmessages":
			receiverID := strings.Split(data.MessageInfo.ReceiverID, "-")
			data.MessageInfo.ReceiverID = receiverID[len(receiverID)-1]
			err := newMessage(&db, ConnectedUsers, data.MessageInfo.SenderID, data.MessageInfo.ReceiverID, data.MessageInfo.Content)
			if err != nil {
				textErr = `{"success": false,  "datas":"newmessage", "error":"` + err.Error() + `"}`
				conn.WriteMessage(websocket.TextMessage, []byte(textErr))
				continue
			}
		case "getmessages":
			receiverID := strings.Split(data.MessageInfo.ReceiverID, "-")
			_, err := getMessage(&db, conn, data.MessageInfo.SenderID, receiverID[len(receiverID)-1])
			if err != nil {
				textErr = `{"success": false,  "datas":"conversation", "error":"` + err.Error() + `"}`
				conn.WriteMessage(websocket.TextMessage, []byte(textErr))
				continue
			}
		}
	}
	log.Println("Client disconected")
	id := deleteConnectedUser(userID)
	broadcastOnlineUsers(ConnectedUsers, "offline", id)
}

func getMessage(db *DataBase, conn *websocket.Conn, UserId string, ReceiverId string) ([]MessageInfo, error) {
	message, err := db.GetMessages(conn, UserId, ReceiverId)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func newMessage(db *DataBase, online map[string]*websocket.Conn, UserId string, ReceiverId string, message string) error {
	err := db.SaveMessage(db.DB, online, UserId, ReceiverId, message)
	if err != nil {
		return err
	}
	return nil
}

func newCommentRequest(db *DataBase, UserId string, PubId string, Comment string) error {
	log.Println("Inserting new comment:", UserId, PubId, Comment) // Add log here
	err := db.InsertNewComment(UserId, PubId, Comment)
	if err != nil {
		return err
	}
	return nil
}

func registerRequest(db *DataBase, username, firstName, lastName, gender, mail, password string) error {
	err := db.InsertUser(username, firstName, lastName, gender, mail, password)
	if err != nil {
		return err
	}
	return nil
}

func loginRequest(db *DataBase, nickname, password string) (int, error) {
	id, err := db.CheckUser(nickname, password)
	return id, err
}

func postRequest(db *DataBase, userId string, categorie string, title string, content string) error {
	err := db.InsertPost(userId, categorie, title, content)
	if err != nil {
		return err
	}
	return nil
}

func broadcastOnlineUsers(online map[string]*websocket.Conn, mode string, id string) error {
	var onlineUser, status string
	for id := range online {
		onlineUser += id + " "
	}

	if mode == "online" {
		status = "online"
	} else {
		status = "offline"
		onlineUser = id
	}
	response := Response{
		DataName: status,
		Data:     onlineUser,
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Println("Erreur lors de la conversion des utilisateurs en JSON:", err)
		return err
	}
	for _, wsconnOfUser := range online {
		if err := wsconnOfUser.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Println("BroadCast Erreur lors de l'envoi des données des utilisateurs via WebSocket:", err)
			return err
		}
	}
	log.Println("Connected in broadcast", online)
	return nil
}

func broadcastTypinInProgress(online map[string]*websocket.Conn, sender string, receiver string, lenght int) error {
	onlineForTyping := make(map[string]*websocket.Conn)
	receiv := strings.Split(receiver, "-")
	receiverID := receiv[len(receiv)-1]

	for id, conn := range online {
		if id == receiverID {
			onlineForTyping[id] = conn
		}
	}

	response := ResponseForTypin{
		DataName: "typing",
		Data:     sender,
		Leng:     lenght,
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Println("Erreur en JSON D zabi:", err)

		return err
	}
	for _, wsconnOfUser := range onlineForTyping {
		if err := wsconnOfUser.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Println("BroadCast Erreur d zab to web:", err)
			return err
		}
	}
	return nil
}
