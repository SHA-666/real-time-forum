package serveur

type RegisterData struct {
	Nickname  string `json:"nickname"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}
type LoginData struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}
type CreatPostData struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Content  string `json:"content"`
}
type UserConnectionData struct {
	Id string `json:"id"`
}
type DataStruct struct {
	DataName     string             `json:"data,omitempty"`
	LoginData    LoginData          `json:"loginData,omitempty"`
	RegisterData RegisterData       `json:"registerData,omitempty"`
	CreatPost    CreatPostData      `json:"creatPostData,omitempty"`
	NewComment   CommentData        `json:"commentaire,omitempty"`
	UserConn     UserConnectionData `json:"userConnection,omitempty"`
	IdOfUser     Profil             `json:"IdOfUser,omitempty"`
	MessageInfo  MessageInfo        `json:"messageInfo,omitempty"`
	Typing       Typings            `json:"typing,omitempty"` // Ensure this matches
}

type Typings struct {
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	Leng int `json:"len"`
}
type MessageInfo struct {
	SenderID     string `json:"sender_id"`
	ReceiverID   string `json:"receiver_id"`
	Content      string `json:"content"`
	SentAt       string `json:"sent_at"`
	SenderName   string `json:"sendername"`
	ReceiverName string `json:"receivername"`
}
type Profil struct {
	UserId string `json:"UserId"`
}
type CommentData struct {
	User    string `json:"user"`
	Comment string `json:"comment"`
	IDPub   string `json:"idpub"`
}

type ResponseForTypin struct {
	DataName string      `json:"datas"`
	Data     interface{} `json:"data"`
	Leng int `json:"len"`
}
type Response struct {
	DataName string      `json:"datas"`
	Data     interface{} `json:"data"`
}
type Pub struct {
	Id           string     `json:"idpub"`
	UserNickname string     `json:"user"`
	Categories   string     `json:"categories"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	CreationDate string     `json:"creation_date"`
	Commentaire  []Comments `json:"comments"`
}
type Comments struct {
	Username     string `json:"username"`
	Content      string `json:"content"`
	CreationDate string `json:"creation_date"`
}
