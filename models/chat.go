package models

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newLine = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	conn   *websocket.Conn
	User   User               `json:"user"`
	ChatID primitive.ObjectID `json:"chat_id"`
	UserID primitive.ObjectID `json:"user_id"`
}

var clients = make(map[*Client]bool)
var rooms = make(map[string]map[*Client]bool)

type Chat struct {
	ID          primitive.ObjectID `bson:"_id, omitempty" json:"id" xml:"id"`
	Participant []Participant      `json:"participants"`
	CreatedAt   time.Time          `bson:"createt_at" json:"createdAt" xml:"createdAt"`
	DeletedAt   time.Time          `bson:"deleted_at" json:"deletedAt"  xml:"deletedAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt" xml:"updatedAt"`
}

type Participant struct {
	ID     primitive.ObjectID `bson:"_id, omitempty" json:"id" xml:"id"`
	ChatID primitive.ObjectID `bson:"chat_id, omitempty" json:"chatId" xml:"chatId"`
	UserID primitive.ObjectID `bson:"user_id" json:"userId" xml:"userId"`
}

type Message struct {
	ID        primitive.ObjectID `bson:"_id, omitempty" json:"id" xml:"id"`
	ChatID    primitive.ObjectID `bson:"chat_id, omitempty" json:"chatId" xml:"chatId"`
	SenderID  primitive.ObjectID `bson:"sender_id" json:"senderId" xml:"senderId"`
	Sender    User               `json:"sender" xml:"sender"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt" xml:"createdAt"`
	DeletedAt time.Time          `bson:"deleted_at" json:"deletedAt"  xml:"deletedAt"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt" xml:"updatedAt"`
	Deleted   bool               `bson:"deleted" json:"deleted" xml:"deleted"`
	Content   string             `bson:"content" json:"content" xml:"content"`
}

func (message *Message) Prepare() {
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()
	message.Deleted = false
}

func (message Message) SaveMessage() (Message, error) {
	ctx := context.TODO()
	db := getMessageCollection()
	message.ID = primitive.NewObjectID()
	message.Prepare()
	_, err := db.InsertOne(ctx, message)
	if err != nil {
		return Message{}, errors.New(fmt.Sprintf("Mesaj kayıt edilirken hata oluştu hata şu şekilde \"%v\"", err))
	}
	return message, nil
}
func (participant Participant) SaveParticipant() (Participant, error) {
	ctx := context.TODO()
	db := getMessageCollection()
	participant.ID = primitive.NewObjectID()
	_, err := db.InsertOne(ctx, participant)
	if err != nil {
		return Participant{}, errors.New(fmt.Sprintf("Katılımcı kayıt edilirken hata oluştu hata şu şekilde \"%v\"", err))
	}
	return participant, nil
}

func (message Message) FindMessagesByChatID(id primitive.ObjectID) ([]Message, error) {
	ctx := context.TODO()
	db := getMessageCollection()
	filter := bson.M{"chat_id": id}
	cursor, err := db.Find(ctx, filter)
	if err != nil {
		return []Message{}, errors.New(fmt.Sprintf("Mesaj  çekilirken hata oluştu hata şu şekilde \"%v\"", err))
	}

	var messages []Message

	if err = cursor.All(ctx, &messages); err != nil {
		return []Message{}, errors.New(fmt.Sprintf("Mesajlar  dönüştülürken hata oluştu hata şu şekilde \"%v\"", err))
	}
	return messages, nil
}

func (chat Chat) NewChat() (Chat, error) {
	ctx := context.TODO()
	db := getMessageCollection()
	chat.ID = primitive.NewObjectID()
	_, err := db.InsertOne(ctx, chat)
	if err != nil {
		return Chat{}, errors.New(fmt.Sprintf("Mesaj kayıt edilirken hata oluştu hata şu şekilde \"%v\"", err))
	}
	return chat, nil
}

func (chat Chat) FindChatByParticipantId(id primitive.ObjectID) ([]Chat, error) {
	ctx := context.TODO()
	db := getMessageCollection()
	filter := bson.M{
		"$or": bson.A{
			bson.M{"subscriber_id": id},
			bson.M{"owner_id": id},
		},
	}
	cursor, err := db.Find(ctx, filter)
	var chats []Chat
	if err != nil {
		return []Chat{}, errors.New(fmt.Sprintf("Sohbeter  çekilirken hata oluştu hata şu şekilde \"%v\"", err))
	}

	if err = cursor.All(ctx, &chats); err != nil {
		return []Chat{}, errors.New(fmt.Sprintf("Mesajlar  dönüştülürken hata oluştu hata şu şekilde \"%v\"", err))
	}
	return chats, nil

}

func (chat Chat) FindChatByID(id primitive.ObjectID) (Chat, error) {
	ctx := context.TODO()
	db := getMessageCollection()
	err := db.FindOne(ctx, bson.M{"_id": id}).Decode(&chat)
	if err != nil {
		return Chat{}, err
	}

	return chat, nil

}

func (chat *Chat) FindAll() ([]Chat, error) {
	ctx := context.TODO()
	db := getMessageCollection()

	cursor, err := db.Find(ctx, bson.D{})
	if err != nil {
		return []Chat{}, err
	}

	var chats []Chat
	if err = cursor.All(ctx, &chats); err != nil {
		return []Chat{}, err
	}
	return chats, nil
}

func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err})
		return
	}
	defer conn.Close()

	client := &Client{conn: conn}
	clients[client] = true

	fmt.Println("Client connected")

	client.User.UserName = "user" + fmt.Sprint(len(clients))
	client.conn.WriteJSON(Message{Content: "Welcome"})

	for {
		var msg Message
		err := client.conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading JSON:", err)
			delete(clients, client)
			break
		}
		sendMessage(client, msg)
	}
}

func JoinRoom(c *gin.Context) {
	ChatID := c.Param("chatID")
	roomId, err := primitive.ObjectIDFromHex(ChatID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": err})
		return
	}
	cht := Chat{}
	_, err = cht.FindChatByID(roomId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": err})
		return
	}
	userID := c.Request.Header.Get("userId")
	uid, err := primitive.ObjectIDFromHex(userID)
	/*ownerId := chat.OwnerID
	subscriberId := chat.SubscriberID
	if uid != subscriberId && uid != ownerId {
		c.JSON(http.StatusUnauthorized, gin.H{})
		fmt.Println(uid != subscriberId && uid != ownerId)
		return
	}*/
	if _, exists := rooms[ChatID]; !exists {
		rooms[ChatID] = make(map[*Client]bool)
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err})
		return
	}
	defer conn.Close()
	client := &Client{conn: conn, UserID: uid, ChatID: roomId}
	clients[client] = true
	rooms[ChatID][client] = true
	user, err := client.User.FindById(uid)
	client.User.UserName = user.UserName
	client.conn.WriteJSON(Message{Content: "Welcome"})
	for {
		var msg Message
		err := client.conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading JSON:", err)
			delete(clients, client)
			break
		}
		sendToRoomMessage(client, msg)
	}
	fmt.Printf("Client joined room %s\n", ChatID)
}

func sendToRoomMessage(sender *Client, msg Message) error {
	fmt.Printf("[%s][%s]: %s\n", sender.User.UserName, sender.ChatID, msg.Content)
	msg.ChatID = sender.ChatID
	msg.SenderID = sender.UserID
	user, err := msg.Sender.FindById(sender.UserID)
	if err != nil {
		return err
	}
	msg.Sender = user
	message, _ := msg.SaveMessage()
	for client := range rooms[string(sender.ChatID.Hex())] {
		if client != sender {
			err := client.conn.WriteJSON(message)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func sendMessage(sender *Client, msg Message) {
	fmt.Printf("[%s]: %s\n", sender.User.UserName, msg.Content)
	for client := range clients {
		if client != sender {
			err := client.conn.WriteJSON(msg)
			if err != nil {
				fmt.Println("Error writing JSON:", err)
				return
			}
		}
	}
}
