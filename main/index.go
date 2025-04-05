package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"

	"learnLoop/main/config"
	"learnLoop/main/models"
	"learnLoop/main/websocket"
)

//go:embed data/quiz1.json
var quiz1Model string

var quiz1 *models.Quiz

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func clientCloseHandler(client *websocket.Client) error {
	// send close message of this user to all users
	err := websocket.SendMessageToAllClients(client.Hub, models.UserDisconnectMessage{
		Envelope: &models.Envelope{
			Type: models.MESSAGE_TYPE_USER_DISCONNECTED,
		},
		From: models.Recipient{
			Type: models.RECIPIENT_TYPE_LEARNER,
			Id:   client.User.Login,
		},
		To: []models.Recipient{
			{
				Type: models.RECIPIENT_TYPE_SESSION,
				Id:   client.User.SessionID,
			},
		},
	})
	if err != nil {
		log.Printf("Error sending UserDisconnectMessage: %v\n", err)
	}
	return err
}

var commandServices = models.CommandServices{
	MessageSender: func(user *models.User, message interface{}) error {
		return websocket.SendMessageToAllClients(hub, message)
	},

	SendUserConnectMessageForAllUsersInSession: func(session *models.Session) error {
		for _, user := range hub.GetUsersInSession(session.SessionID) {
			err := websocket.SendMessageToAllClients(hub, models.UserConnectMessage{
				Envelope: &models.Envelope{
					Type: models.MESSAGE_TYPE_USER_CONNECTED,
				},
				From: models.Recipient{
					Type: models.RECIPIENT_TYPE_LEARNER,
					Id:   user.Login,
				},
				To: []models.Recipient{
					{
						Type: models.RECIPIENT_TYPE_SESSION,
						Id:   user.SessionID,
					},
				},
			})
			if err != nil {
				log.Printf("Error sending UserConnectMessage for user %s: %v\n", user.Login, err)
			}
		}
		return nil
	},

	GetQuiz: func(quizId int) (*models.Quiz, error) {
		if quizId != 1 {
			return nil, fmt.Errorf("unknown quiz ID: %d", quizId)
		}
		return quiz1, nil
	},
}

func clientMessageHandler(client *websocket.Client, message []byte) error {
	command, err := models.ParseCommand(message)
	if err != nil {
		return fmt.Errorf("error parsing message: %v", err)
	}
	if command != nil {
		err := (*command).Execute(
			client.User,
			client.Hub.GetSession(client.User.SessionID),
			commandServices,
		)
		if err != nil {
			return fmt.Errorf("error executing command: %v", err)
		}
	}
	return err
}

var hub *websocket.Hub = nil

func main() {
	config.Init()

	// Parse the embedded quiz JSON
	var err error
	quiz1, err = models.ParseQuiz(quiz1Model)
	if err != nil {
		log.Fatalf("Failed to parse quiz1: %v", err)
	}
	log.Printf("Loaded quiz: %s with %d questions", quiz1.Title, len(quiz1.Questions))

	hub = websocket.NewHub()
	if config.DevMode {
		log.Printf("Starting server on port %s in dev mode\n", config.Addr)
	} else {
		log.Printf("Starting server on port %s in production mode\n", config.Addr)
	}
	go hub.Run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r, clientCloseHandler, clientMessageHandler)
	})
	err = http.ListenAndServe(config.Addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
