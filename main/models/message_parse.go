package models

import (
	"encoding/json"
	"fmt"
	"log"
)

// createCommand is a factory function that creates an Command based on the envelope type
func createCommand(envelope Envelope) (Command, error) {
	switch envelope.Type {
	case MESSAGE_TYPE_MESSAGE:
		return &Message{}, nil
	case MESSAGE_TYPE_USER_CONNECTED:
		return &UserConnectMessage{}, nil
	case MESSAGE_TYPE_USER_DISCONNECTED:
		return &UserDisconnectMessage{}, nil
	case MESSAGE_TYPE_QUIZ_MESSAGE:
		if envelope.Action == QUIZ_MESSAGE_ACTION_START {
			return &QuizStartMessage{}, nil
		} else if envelope.Action == QUIZ_MESSAGE_ACTION_QUESTION_START {
			return &QuizQuestionMessage{}, nil
		} else if envelope.Action == QUIZ_MESSAGE_ACTION_LEARNER_ANSWER {
			return &QuizLearnerAnswerMessage{}, nil
		} else if envelope.Action == QUIZ_MESSAGE_ACTION_LEARNER_ANSWER_FREE_TEXT {
			return &QuizLearnerAnswerFreeTextMessage{}, nil
		} else if envelope.Action == QUIZ_MESSAGE_ACTION_QUESTION_STATS {
			return &QuizQuestionStatsMessage{}, nil
		} else if envelope.Action == QUIZ_MESSAGE_ACTION_QUESTION_END {
			return &QuizQuestionEndMessage{}, nil
		} else if envelope.Action == QUIZ_MESSAGE_ACTION_NEXT_QUESTION {
			return &QuizNextQuestionMessage{}, nil
		} else {
			return nil, fmt.Errorf("unknown quiz message action: %d", envelope.Action)
		}
	default:
		return nil, fmt.Errorf("unknown message type: %d", envelope.Type)
	}
}

// ParseCommand checks if the message represents a JsonMessage.
func ParseCommand(message []byte) (*Command, error) {
	messageStr := string(message)
	log.Printf("Parsing message: %s\n", messageStr)

	// Try to parse as JSON
	var envelope Envelope
	err := json.Unmarshal(message, &envelope)
	if err != nil {
		log.Printf("Error parsing message as JSON: %v\n", err)
		return nil, err
	}

	// Create the appropriate Action based on the envelope type
	command, err := createCommand(envelope)
	if err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}

	// Unmarshal the full message into the command
	if err := json.Unmarshal(message, command); err != nil {
		log.Printf("Error unmarshaling message: %v\n", err)
		return nil, err
	}

	return &command, nil
}
