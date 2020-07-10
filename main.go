package main

import (
	"context"
	"log"
	"net/http"

	"github.com/camilocorreaUdeA/WebChatApp/utility"
	socketio "github.com/googollee/go-socket.io"
)

var loggedUsers []string

func main() {

	context := context.Background()
	url := "mongodb://localhost:27017"
	dbname := "webchatapp"
	colname := "users"

	// Connection to MongoDB - Database: webchatapp
	client, database := utility.DataBaseConnection(context, url, dbname)

	defer client.Disconnect(context)

	//WebChatApp subscribers collection
	usersCollection := utility.DataBaseCollection(database, colname)

	var signedUsers []utility.SignedUser

	//Store collection in local struct for easier management
	utility.GetAllCollectionData(context, usersCollection, &signedUsers)

	//Start Socket.IO server (Underlying web sockets)
	server, err := socketio.NewServer(nil)

	if err != nil {
		log.Fatal(err)
	}

	//Response to client socket connection event
	server.OnConnect("/", func(so socketio.Conn) error {
		so.Join("chat_room")
		return nil
	})

	//Response to client socket disconnection event (close browser/browser tab)
	server.OnDisconnect("/", func(so socketio.Conn, reason string) {
		log.Println("User with ID: ", so.ID(), " disconnected")
	})

	//Response to client socket login event:
	//Check subscribers collection
	//Check credentials provided by user
	//Join chat or ask to subscribe
	server.OnEvent("/", "authentication", func(so socketio.Conn, msg string) {

		var user utility.User

		user.Username, user.Password = utility.SplitDataString(msg)

		var authorizedUsers []utility.SignedUser

		utility.GetAllCollectionData(context, usersCollection, &authorizedUsers)

		for _, logged := range loggedUsers {
			if logged == user.Username {
				so.Emit("in_session", user.Username)
				return
			}
		}

		for _, authUser := range authorizedUsers {
			if (authUser.Username == user.Username) && (authUser.Password == user.Password) {
				loggedUsers = append(loggedUsers, user.Username)
				so.Emit("logged", user.Username, loggedUsers)
				return
			}
		}

		so.Emit("invalid_credentials", user.Username)
	})

	//Response to client socket signin event
	//Insert new subscriber data into database
	server.OnEvent("/", "signing", func(so socketio.Conn, msg string) {

		var user utility.User
		user.Username, user.Password = utility.SplitDataString(msg)

		for _, signedUser := range signedUsers {
			if signedUser.Username == user.Username {
				so.Emit("used_username", user.Username)
				return
			}
		}

		newUser := utility.SignedUser{
			Username: user.Username,
			Password: user.Password,
		}

		err := utility.InsertDataCollection(context, usersCollection, newUser)
		if err != nil {
			so.Emit("error_signin")
		}

		so.Emit("signed_user", user.Username)
	})

	//Response to client socket chat message event:
	//Check either public or private message
	//Broadcast message to chat room
	server.OnEvent("/", "chat message", func(so socketio.Conn, msg string, user string, receiver string, pmsg bool) {
		if pmsg {
			msg = "Private message: " + msg
		}
		server.BroadcastToRoom("", "chat_room", "chat message", msg, user, receiver, pmsg, loggedUsers)
	})

	//Response to client socket leave chat room event (User pressed leave button):
	//Remove user from currently logged users list
	server.OnEvent("/", "leave_room", func(so socketio.Conn, user string) {

		updtList, err := utility.Remove(loggedUsers, user)

		if err != nil {
			log.Println(err)
		}

		loggedUsers = updtList
		so.Emit("user_removed", loggedUsers)
	})

	//Disconnect client socket after user left chat room
	server.OnEvent("/", "user_left", func(so socketio.Conn) {
		log.Println("User with ID: ", so.ID(), " disconnected")
		so.Leave("chat_room")
	})

	//Goroutine to serve client sockets
	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./public")))
	log.Println("Server on port 3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
