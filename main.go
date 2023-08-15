package main

import (
	"chatapp/controllers"
	"chatapp/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	//Url için api linkleri

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	//config.AllowOrigins = []string{"http://localhost:3001"}
	router.Use(cors.New(config))

	//Kullanıcılar için api linkleri
	user := router.Group("user")
	user.POST("/register", controllers.CreateUser)
	user.GET("/getAll", controllers.GetAllUsers)
	user.GET("/:userId", controllers.GetUserById)

	chat := router.Group("chat")
	chat.POST("/newmessage", controllers.NewMessage)
	chat.GET("/getbyChatId/:chatId", controllers.GetMessagesByChatID)
	chat.POST("/newchat", controllers.NewChat)
	//chat.GET("/getbyparticipant/:userId", controllers.GetChatByParticipantID)
	chat.GET("/getbyId/:id", controllers.GetChatByID)
	chat.GET("/getall", controllers.GetAllChats)

	chat.GET("/ws", func(ctx *gin.Context) {
		models.HandleWebSocket(ctx)
	})
	chat.GET("/room/:chatID", func(ctx *gin.Context) {
		models.JoinRoom(ctx)
	})

	router.Run("localhost:8080")
}
