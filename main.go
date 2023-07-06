package main

import (
	"log"

	"github.com/gin-gonic/gin"

	socketio "github.com/googollee/go-socket.io"
)

type Message struct {
	ChatId  string `form:"chat_id" json:"chat_id"`
	Message string `form:"message" json:"message"`
}

func main() {
	router := gin.New()

	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		log.Println("connected:", s.ID())
		return nil
	})

	// Listen message
	server.OnEvent("/chat", "message", func(s socketio.Conn, msg Message) {
		s.Emit(msg.ChatId, msg)
	})

	server.OnEvent("/", "msg", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		return "recv " + msg
	})

	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("closed", reason)
	})

	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()

	defer server.Close()

	router.GET("/socket.io/*any", gin.WrapH(server))
	router.POST("/socket.io/*any", gin.WrapH(server))
	// router.StaticFS("/public", http.Dir("../asset"))

	router.POST("/message", func(c *gin.Context) {
		var message Message
		c.Bind(&message)

		server.BroadcastToNamespace("/chat", message.ChatId, message)

		c.JSON(200, message)
	})

	if err := router.Run(":8000"); err != nil {
		log.Fatal("failed run app: ", err)
	}
}
