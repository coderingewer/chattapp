package models

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	ChatID  string
	message chan *Message
}

func (client *Client) Prepare() {
	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
}

var clients = make(map[*Client]bool)
var rooms = make(map[string]map[*Client]bool)

func JoinRoom(c *gin.Context) {
	//gerekli parametreler alınır
	ChatID := c.Param("chatId")
	userID := c.Request.Header.Get("userId")
	uid, err := primitive.ObjectIDFromHex(userID)
	if _, exists := rooms[ChatID]; !exists {
		rooms[ChatID] = make(map[*Client]bool)
	}
	cid, err := primitive.ObjectIDFromHex(ChatID)
	if _, exists := rooms[ChatID]; !exists {
		rooms[ChatID] = make(map[*Client]bool)
	}
	//Kullanıcıların bu sohbete ait olup olmadığı kontrol edilir
	cht := Chat{}
	chat, err := cht.FindChatByID(cid)
	fmt.Println(cid)
	switch {
	case uid == chat.Participants[0].UserID:
		fmt.Println("uid")
	case uid == chat.Participants[1].UserID:
		fmt.Println("huh")
	default:
		fmt.Println("hata df")
		return
	}

	//http bağlantısı ws bağlantısına yükseltilir
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err})
		return
	}
	defer conn.Close()
	//istemci bağlanır
	client := &Client{conn: conn, ChatID: ChatID}
	clients[client] = true
	rooms[ChatID][client] = true
	client.conn.WriteJSON(Message{Content: "Welcome"})
	client.Prepare()
	//mesajları almak için gerekli olan işlemler yapılır
	for {
		usr := User{}
		messageSender, err := usr.FindById(uid)
		if err != nil {
			fmt.Println("[ERROR]", err, "while finding user by id", uid)
		}
		msg := Message{SenderID: uid, ChatID: cid, Sender: messageSender}
		err = client.conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading JSON:", err)
			client.hub.unregister <- client
			break
		}
		SendMessageToRoom(client, msg)
	}
}

// Odaya mesaj göndermeye yarayan fonksiyon
func SendMessageToRoom(sender *Client, msg Message) error {
	message, err := msg.SaveMessage()
	if err != nil {
		return err
	}
	for client := range rooms[string(sender.ChatID)] {
		if client != sender {
			err := client.conn.WriteJSON(message)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
