package controllers

import (
	"chatapp/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewMessage(c *gin.Context) {
	message := models.Message{}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": http.StatusUnprocessableEntity})
		return
	}

	err = json.Unmarshal(body, &message)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": http.StatusUnprocessableEntity})
		return
	}

	messageSaved, err := message.SaveMessage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messageSaved)
}
func NewChat(c *gin.Context) {
	chat := models.Chat{}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": http.StatusUnprocessableEntity})
		return
	}

	err = json.Unmarshal(body, &chat)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": http.StatusUnprocessableEntity})
		return
	}

	chatSaved, err := chat.NewChat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": err.Error()})
		return
	}
	userID := c.Request.Header.Get("userId")
	recieverID := c.Request.Header.Get("recieverID")
	ids := [2]string{userID, recieverID}
	for _, v := range ids {
		participant := models.Participant{}
		uid, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": err})
			return
		}
		participant.UserID = uid
		participant.ChatID = chatSaved.ID
		participantSaved, err := participant.SaveParticipant()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": err})
			return
		}
		chatSaved.Participant = append(chatSaved.Participant, participantSaved)
	}
	c.JSON(http.StatusOK, chatSaved)
}

func GetMessagesByChatID(c *gin.Context) {
	message := models.Message{}
	id := c.Param("chatId")
	chatId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "geçersiz ID"})
		return
	}
	messages, err := message.FindMessagesByChatID(chatId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": fmt.Sprintf("%v", err)})
		return
	}
	c.JSON(http.StatusOK, messages)
}

/*
	func GetChatByParticipantID(c *gin.Context) {
		chat := models.Chat{}
		id := c.Param("userId")
		participantID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "geçersiz ID"})
			return
		}
		chats, err := chat.FindChatByParticipantId(participantID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": fmt.Sprintf("%v", err)})
			return
		}
		c.JSON(http.StatusOK, chats)
	}
*/
func GetChatByID(c *gin.Context) {
	chat := models.Chat{}
	id := c.Param("id")
	chatId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "geçersiz ID"})
		return
	}
	chats, err := chat.FindChatByID(chatId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": fmt.Sprintf("%v", err)})
		return
	}
	c.JSON(http.StatusOK, chats)
}

func GetAllChats(c *gin.Context) {
	chat := models.Chat{}
	chats, err := chat.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chats)
}
