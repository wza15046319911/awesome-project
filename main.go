package main

import (
	api "awesome/api"
	"fmt"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"log"
)

var client, _ = initialDatabase()
var database = client.Database("COMP90018")
var eventApi = &api.EventApi{
	Database:   *database,
	Collection: *database.Collection("Events"),
}

var userApi = &api.UserApi{
	Database:          *database,
	UserCollection:    *database.Collection("Users"),
	ProfileCollection: *database.Collection("Profile"),
}

func main() {
	router := gin.New()

	server := socketio.NewServer(nil)
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		s.Emit("connect", "connected")
		fmt.Println("connected:", s.ID())
		return nil
	})
	server.OnEvent("/", "join", func(s socketio.Conn, msg string) {
		s.Join(msg)

	})
	server.OnEvent("/", "ping", func(s socketio.Conn, msg string) {
		s.Emit("pong", msg)
	})
	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})
	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()
	defer server.Close()
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})
	router.GET("/socket.io/*any", gin.WrapH(server))
	router.POST("/socket.io/*any", gin.WrapH(server))
	// root group
	//root := router.Group("/api/v1")
	//{
	//	// user api
	//	root.POST("/login", userApi.Login)
	//	root.POST("/users", userApi.RegisterUser)
	//	root.GET("/users/profile", userApi.GetProfile)
	//	root.POST("/users/profile", userApi.UpdateProfile)
	//	root.GET("/users/avatars", userApi.GetAvatars)
	//
	//	// event api
	//	root.GET("/events", eventApi.GetEvent)
	//	root.POST("/events", eventApi.AddEvent)
	//	root.PATCH("/events", eventApi.UpdateEvent)
	//	root.POST("/events/delete", eventApi.DeleteEvent)
	//}

	if err := router.Run("localhost:8000"); err != nil {
		log.Fatal("failed run app: ", err)
	}
}
