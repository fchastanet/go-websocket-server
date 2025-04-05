package models

import (
	"encoding/json"
	"testing"
)

func TestEnvelopeUnmarshal(t *testing.T) {
	jsonData := `{"type": 2, "timestamp": 123456789}`

	var envelope Envelope
	err := json.Unmarshal([]byte(jsonData), &envelope)
	if err != nil {
		t.Fatalf("Failed to unmarshal Envelope: %v", err)
	}

	if envelope.Type != MESSAGE_TYPE_MESSAGE {
		t.Errorf("Expected type %d, got %d", MESSAGE_TYPE_MESSAGE, envelope.Type)
	}

	if envelope.Timestamp != 123456789 {
		t.Errorf("Expected timestamp %d, got %d", 123456789, envelope.Timestamp)
	}
}

func TestMessageUnmarshal(t *testing.T) {
	jsonData := `{
		"type": 2,
		"from": {"type": 0, "id": "user1"},
		"to": [{"type": 1, "id": "user2"}, {"type": 0, "id": "user3"}],
		"msg": "Hello, World!",
		"timestamp": 123456789
	}`

	var message Message
	err := json.Unmarshal([]byte(jsonData), &message)
	if err != nil {
		t.Fatalf("Failed to unmarshal Message: %v", err)
	}

	// Verify the envelope fields were correctly unmarshaled
	if message.Envelope == nil {
		t.Fatalf("Envelope is nil")
	}

	if message.Type != MESSAGE_TYPE_MESSAGE {
		t.Errorf("Expected type %d, got %d", MESSAGE_TYPE_MESSAGE, message.Type)
	}

	// Check sender
	if message.From.Id != "user1" || message.From.Type != RECIPIENT_TYPE_SESSION {
		t.Errorf("Incorrect sender: %+v", message.From)
	}

	// Check recipients
	if len(message.To) != 2 {
		t.Errorf("Expected 2 recipients, got %d", len(message.To))
	} else {
		if message.To[0].Id != "user2" || message.To[0].Type != RECIPIENT_TYPE_LEARNER {
			t.Errorf("Incorrect first recipient: %+v", message.To[0])
		}
		if message.To[1].Id != "user3" || message.To[1].Type != RECIPIENT_TYPE_SESSION {
			t.Errorf("Incorrect second recipient: %+v", message.To[1])
		}
	}

	// Check message content
	if message.Msg != "Hello, World!" {
		t.Errorf("Expected message content 'Hello, World!', got '%s'", message.Msg)
	}
}

func TestUserConnectMessageUnmarshal(t *testing.T) {
	jsonData := `{
		"type": 0,
		"from": {"type": 0, "id": "user1"},
		"to": [{"type": 0, "id": "user2"}]
	}`

	var message UserConnectMessage
	err := json.Unmarshal([]byte(jsonData), &message)
	if err != nil {
		t.Fatalf("Failed to unmarshal UserConnectMessage: %v", err)
	}

	if message.Type != MESSAGE_TYPE_USER_CONNECTED {
		t.Errorf("Expected type %d, got %d", MESSAGE_TYPE_USER_CONNECTED, message.Type)
	}

	if message.From.Id != "user1" {
		t.Errorf("Expected from.id 'user1', got '%s'", message.From.Id)
	}
}
