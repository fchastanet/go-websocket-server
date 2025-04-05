package models

import (
	"fmt"
	"testing"
)

func TestParseMessage(t *testing.T) {
	tests := []struct {
		name          string
		messageJSON   string
		expectedType  string
		expectedError bool
	}{
		{
			name: "Valid Regular Message",
			messageJSON: `{
				"type": 2,
				"from": {"type": 0, "id": "user1"},
				"to": [{"type": 0, "id": "user2"}],
				"msg": "Hello, World!"
			}`,
			expectedType:  "*models.Message",
			expectedError: false,
		},
		{
			name: "Valid User Connect Message",
			messageJSON: `{
				"type": 0,
				"from": {"type": 0, "id": "user1"},
				"to": [{"type": 0, "id": "user2"}]
			}`,
			expectedType:  "*models.UserConnectMessage",
			expectedError: false,
		},
		{
			name: "Valid User Disconnect Message",
			messageJSON: `{
				"type": 1,
				"from": {"type": 0, "id": "user1"},
				"to": [{"type": 0, "id": "user2"}]
			}`,
			expectedType:  "*models.UserDisconnectMessage",
			expectedError: false,
		},
		{
			name:          "Invalid JSON",
			messageJSON:   `{"type": 2, "from":}`,
			expectedType:  "",
			expectedError: true,
		},
		{
			name: "Message with Additional Fields",
			messageJSON: `{
				"type": 2,
				"from": {"type": 0, "id": "user1"},
				"to": [{"type": 0, "id": "user2"}],
				"msg": "Hello, World!",
				"timestamp": 123456789
			}`,
			expectedType:  "*models.Message",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, err := ParseCommand([]byte(tt.messageJSON))

			// Check if error matches expected
			if (err != nil) != tt.expectedError {
				t.Errorf("ParseCommand() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			// If no error is expected, validate the result type
			if err == nil {
				if action == nil {
					t.Errorf("ParseCommand() returned nil action for: %s", tt.name)
					return
				}

				// Check the concrete type of the action
				actualType := GetTypeName(*action)
				if actualType != tt.expectedType {
					t.Errorf("ParseCommand() returned type %s, want %s", actualType, tt.expectedType)
				}

				// Validate specific fields based on the message type
				validateMessageFields(t, *action, tt.messageJSON)
			}
		})
	}
}

// Test for handling unknown message types
func TestParseMessage_UnknownType(t *testing.T) {
	// This should return an error due to unknown message type
	_, err := ParseCommand([]byte(`{"type": 99}`))
	if err == nil {
		t.Errorf("ParseCommand() with unknown message type should return an error")
	}
}

// GetTypeName returns the type name as a string
func GetTypeName(i interface{}) string {
	if i == nil {
		return ""
	}
	// Use %T to get the type including package path
	return fmt.Sprintf("%T", i)
}

// Helper to validate fields of specific message types
func validateMessageFields(t *testing.T, action Command, jsonStr string) {
	switch msg := action.(type) {
	case *Message:
		if msg.Msg == "" {
			t.Errorf("Message has empty msg field")
		}
	case *UserConnectMessage:
		if msg.From.Id == "" {
			t.Errorf("UserConnectMessage has empty from.id field")
		}
	case *UserDisconnectMessage:
		if msg.From.Id == "" {
			t.Errorf("UserDisconnectMessage has empty from.id field")
		}
	}
}
