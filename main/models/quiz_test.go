package models

import (
	"testing"
)

// Test data
var validQuizJSON = `{
	"id": 1,
	"title": "Test Quiz",
	"url": "/quiz/1",
	"questions": [
		{
			"id": 101,
			"question": "What is Go?",
			"url": "/question/101",
			"answers": [
				{
					"id": 1001,
					"title": "A programming language",
					"url": "/answer/1001",
					"correct": 1
				},
				{
					"id": 1002,
					"title": "A board game",
					"url": "/answer/1002",
					"correct": 0
				}
			]
		},
		{
			"id": 102,
			"question": "What year was Go released?",
			"url": "/question/102",
			"answers": [
				{
					"id": 1003,
					"title": "2007",
					"url": "/answer/1003",
					"correct": 0
				},
				{
					"id": 1004,
					"title": "2009",
					"url": "/answer/1004",
					"correct": 1
				}
			]
		}
	]
}`

var invalidQuizJSON = `{
	"id": 1,
	"title": "Broken Quiz",
	missing closing brace
}`

func TestParseQuiz(t *testing.T) {
	t.Run("Valid JSON", func(t *testing.T) {
		quiz, err := ParseQuiz(validQuizJSON)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if quiz == nil {
			t.Fatal("Expected quiz to be parsed, got nil")
		}

		if quiz.ID != 1 {
			t.Errorf("Expected ID 1, got %d", quiz.ID)
		}

		if quiz.Title != "Test Quiz" {
			t.Errorf("Expected title 'Test Quiz', got '%s'", quiz.Title)
		}

		if len(quiz.Questions) != 2 {
			t.Errorf("Expected 2 questions, got %d", len(quiz.Questions))
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		quiz, err := ParseQuiz(invalidQuizJSON)

		if err == nil {
			t.Error("Expected an error, got nil")
		}

		if quiz != nil {
			t.Errorf("Expected quiz to be nil, got %+v", quiz)
		}
	})
}

func TestGetAnswerByID(t *testing.T) {
	quiz, _ := ParseQuiz(validQuizJSON)
	question := quiz.Questions[0]

	t.Run("Existing Answer", func(t *testing.T) {
		answer := question.GetAnswerByID(1001)

		if answer == nil {
			t.Fatal("Expected to find answer, got nil")
		}

		if answer.ID != 1001 {
			t.Errorf("Expected ID 1001, got %d", answer.ID)
		}

		if answer.Title != "A programming language" {
			t.Errorf("Expected title 'A programming language', got '%s'", answer.Title)
		}

		if answer.Correct != 1 {
			t.Errorf("Expected correct=1, got %d", answer.Correct)
		}
	})

	t.Run("Non-existent Answer", func(t *testing.T) {
		answer := question.GetAnswerByID(9999)

		if answer != nil {
			t.Errorf("Expected nil for non-existent answer, got %+v", answer)
		}
	})
}

func TestGetQuestionByID(t *testing.T) {
	quiz, _ := ParseQuiz(validQuizJSON)

	t.Run("Existing Question", func(t *testing.T) {
		question := quiz.GetQuestionByID(101)

		if question == nil {
			t.Fatal("Expected to find question, got nil")
		}

		if question.ID != 101 {
			t.Errorf("Expected ID 101, got %d", question.ID)
		}

		if question.Question != "What is Go?" {
			t.Errorf("Expected question 'What is Go?', got '%s'", question.Question)
		}
	})

	t.Run("Non-existent Question", func(t *testing.T) {
		question := quiz.GetQuestionByID(9999)

		if question != nil {
			t.Errorf("Expected nil for non-existent question, got %+v", question)
		}
	})
}
