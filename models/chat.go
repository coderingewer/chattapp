package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chat struct {
	ID           primitive.ObjectID `bson:"_id, omitempty" json:"id" xml:"id"`
	Participants []Participant      `json:"participants"`
	SpecialID    string             `bson:"special_id" json:"special_id" xml:"special_id"`
	CreatedAt    time.Time          `bson:"createt_at" json:"createdAt" xml:"createdAt"`
	DeletedAt    time.Time          `bson:"deleted_at" json:"deletedAt"  xml:"deletedAt"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updatedAt" xml:"updatedAt"`
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

func (participant Participant) FindParticipantsByChatID(id primitive.ObjectID) ([]Participant, error) {
	ctx := context.TODO()
	db := getMessageCollection()
	filter := bson.M{"chat_id": id}
	cursor, err := db.Find(ctx, filter)
	if err != nil {
		return []Participant{}, errors.New(fmt.Sprintf("Katılımcılar çekilirken hata oluştu hata şu şekilde \"%v\"", err))
	}

	var participants []Participant

	if err = cursor.All(ctx, &participants); err != nil {
		return []Participant{}, errors.New(fmt.Sprintf("Katılımcılar  dönüştülürken hata oluştu hata şu şekilde \"%v\"", err))
	}
	return participants, nil
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
	participant := Participant{}
	participants, err := participant.FindParticipantsByChatID(id)
	chat.Participants = participants
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
	for i, _ := range chats {
		participant := Participant{}
		participants, _ := participant.FindParticipantsByChatID(chats[i].ID)
		chats[i].Participants = participants
	}
	return chats, nil
}

func (chat Chat) FindChatBySpecialID(id string) bool {
	ctx := context.TODO()
	db := getMessageCollection()
	err := db.FindOne(ctx, bson.M{"special_id": id}).Decode(&chat)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (chat *Chat) CheckParticipant(userId, recieverId string) bool {
	ids := []string{userId, recieverId}
	if chat.FindChatBySpecialID(ids[0] + "+" + ids[1]) {
		if chat.FindChatBySpecialID(ids[1] + "+" + ids[0]) {
			return false
		}
		return false
	}
	return true
}
