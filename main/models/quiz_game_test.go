package models

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func newQuizGame() *QuizGame {
	quiz, _ := ParseQuiz(validQuizJSON)
	quizGame := &QuizGame{
		quiz:            quiz,
		StartedBy:       "badLogin",
		StartedAt:       time.Unix(0, 0),
		questionTimeout: 30 * time.Minute,
		GetConnectedPlayersCount: func() int {
			return 2
		},
	}
	quizGame.Start(quiz, &User{Login: "login"})
	return quizGame
}

func TestGetNextQuestion(t *testing.T) {
	quiz, _ := ParseQuiz(validQuizJSON)
	quizGame := QuizGame{
		quiz: quiz,
	}

	t.Run("Has Next Question", func(t *testing.T) {
		nextQuestion := quizGame.getNextQuestion()

		if nextQuestion == nil {
			t.Fatal("Expected to find next question, got nil")
		}

		if nextQuestion.ID != 102 {
			t.Errorf("Expected next question ID 102, got %d", nextQuestion.ID)
		}
	})

	t.Run("Last Question", func(t *testing.T) {
		nextQuestion := quizGame.getNextQuestion()

		if nextQuestion != nil {
			t.Errorf("Expected nil for last question's next, got %+v", nextQuestion)
		}
	})

	t.Run("Non-existent Current Question", func(t *testing.T) {
		nextQuestion := quizGame.getNextQuestion()

		if nextQuestion != nil {
			t.Errorf("Expected nil for non-existent question's next, got %+v", nextQuestion)
		}
	})
}

func TestQuizGameNextQuizQuestion(t *testing.T) {
	quizGame := newQuizGame()
	t.Run("Start Game", func(t *testing.T) {
		if quizGame.StartedBy != "login" {
			t.Errorf("Expected StartedBy to be 'login', got %s", quizGame.StartedBy)
		}
		if quizGame.currentQuestionIndex != -1 {
			t.Errorf("Expected currentQuestionIndex to be -1, got %d", quizGame.currentQuestionIndex)
		}
		if quizGame.quiz == nil {
			t.Errorf("Expected quiz to be set, got nil")
		}
		if quizGame.quiz.ID != 1 {
			t.Errorf("Expected quiz ID to be 1, got %d", quizGame.quiz.ID)
		}
		if quizGame.StartedAt.IsZero() {
			t.Error("Expected quiz started time to be set")
		}
		if len(quizGame.players) != 0 {
			t.Errorf("Expected players to be empty, got %+v", quizGame.players)
		}
		if quizGame.questionStats == nil {
			t.Error("Expected questionStats to be set")
		}
		if len(quizGame.questionStats) != 0 {
			t.Errorf("Expected questionStats to be empty, got %+v", quizGame.questionStats)
		}
		if quizGame.playerStats == nil {
			t.Error("Expected playerStats to be set")
		}
		if len(quizGame.playerStats) != 0 {
			t.Errorf("Expected playerStats to be empty, got %+v", quizGame.playerStats)
		}
		if quizGame.StartedBy != "login" {
			t.Errorf("Expected StartedBy to be 'login', got %s", quizGame.StartedBy)
		}
		if quizGame.currentQuestionIndex != -1 {
			t.Errorf("Expected currentQuestionIndex to be -1, got %d", quizGame.currentQuestionIndex)
		}
	})

	t.Run("Next question => question 1", func(t *testing.T) {
		msg := quizGame.NextQuizQuestionMessage()
		// generate json for message and compare with expected
		msgStr, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Error marshaling JsonMessage: %v\n", err)
		}
		expected := `{"type":4,"action":1,"quizInfo":{"id":1,"title":"Test Quiz","type":0,"url":"/quiz/1","startedBy":"login"},"question":{"id":101,"question":"What is Go?","questionType":0,"url":"/question/101","answers":[{"id":1001,"title":"A programming language","url":"/answer/1001","correct":-1},{"id":1002,"title":"A board game","url":"/answer/1002","correct":-1}]},"questionType":0,"questionNumber":1,"questionCount":2,"timeout":30}`
		if !bytes.Equal(msgStr, []byte(expected)) {
			t.Errorf("Expected message to be %s, got %s", expected, msgStr)
		}
		if quizGame.currentQuestionIndex != 0 {
			t.Errorf("Expected currentQuestionIndex to be 0, got %d", quizGame.currentQuestionIndex)
		}
	})

	t.Run("Next question  => question 2", func(t *testing.T) {
		msg := quizGame.NextQuizQuestionMessage()
		// generate json for message and compare with expected
		msgStr, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Error marshaling JsonMessage: %v\n", err)
		}
		expected := `{"type":4,"action":1,"quizInfo":{"id":1,"title":"Test Quiz","type":0,"url":"/quiz/1","startedBy":"login"},"question":{"id":102,"question":"What year was Go released?","questionType":0,"url":"/question/102","answers":[{"id":1003,"title":"2007","url":"/answer/1003","correct":-1},{"id":1004,"title":"2009","url":"/answer/1004","correct":-1}]},"questionType":0,"questionNumber":2,"questionCount":2,"timeout":30}`
		if !bytes.Equal(msgStr, []byte(expected)) {
			t.Errorf("Expected message to be %s, got %s", expected, msgStr)
		}
	})

	t.Run("Next question  => should be stats", func(t *testing.T) {
		msg := quizGame.NextQuizQuestionMessage()
		// generate json for message and compare with expected
		msgStr, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Error marshaling JsonMessage: %v\n", err)
		}
		expected := `{"type":4,"action":6,"quizId":1,"learnersCount":0,"playerStats":{}}`
		if !bytes.Equal(msgStr, []byte(expected)) {
			t.Errorf("Expected message to be %s, got %s", expected, msgStr)
		}
		if quizGame.currentQuestionIndex != 2 {
			t.Errorf("Expected currentQuestionIndex to be 2, got %d", quizGame.currentQuestionIndex)
		}
	})
}

func TestQuizGameAnsweringQuestion(t *testing.T) {
	quizGame := newQuizGame()
	quizGame.players = []string{"login1", "login2"}

	quizGame.NextQuizQuestionMessage()

	t.Run("Answer first question - bad question id", func(t *testing.T) {
		stats, err := quizGame.AnswerMCQuestion(-1, nil, &User{Login: "login1"})
		if stats != nil {
			t.Errorf("Expected stats to be nil, got %+v", stats)
		}
		if err.Error() != "question ID -1 does not match current question ID 101" {
			t.Errorf("Expected error, got %s", err)
		}
		if quizGame.currentQuestionIndex != 0 {
			t.Errorf("Expected currentQuestionIndex to be 0, got %d", quizGame.currentQuestionIndex)
		}
	})

	t.Run("Answer first question - bad current question id", func(t *testing.T) {
		stats, err := quizGame.AnswerMCQuestion(102, nil, &User{Login: "login1"})
		if stats != nil {
			t.Errorf("Expected stats to be nil, got %+v", stats)
		}
		if err.Error() != "question ID 102 does not match current question ID 101" {
			t.Errorf("Expected error, got %s", err)
		}
		if quizGame.currentQuestionIndex != 0 {
			t.Errorf("Expected currentQuestionIndex to be 0, got %d", quizGame.currentQuestionIndex)
		}
	})

	t.Run("Answer first question - bad answer ids", func(t *testing.T) {
		stats, err := quizGame.AnswerMCQuestion(101, []int{112}, &User{Login: "login1"})
		if stats != nil {
			t.Errorf("Expected stats to be nil, got %+v", stats)
		}
		if err.Error() != "answer ID 112 is invalid" {
			t.Errorf("Expected error, got %s", err)
		}
		if quizGame.currentQuestionIndex != 0 {
			t.Errorf("Expected currentQuestionIndex to be 0, got %d", quizGame.currentQuestionIndex)
		}
	})

	t.Run("Answer first question - good answer", func(t *testing.T) {
		stats, err := quizGame.AnswerMCQuestion(101, []int{1001}, &User{Login: "login1"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if stats == nil {
			t.Fatal("Expected stats, got nil")
		}
		// generate json for message and compare with expected
		msgStr, err := json.Marshal(stats)
		if err != nil {
			t.Errorf("Error marshaling JsonMessage: %v\n", err)
		}
		expected := `{"type":4,"action":3,"questionId":101,"status":0,"learnersCount":2,"answeredCount":1,"answersStats":{"1001":{"answerId":1001,"count":1,"correct":1},"1002":{"answerId":1002,"count":0,"correct":0}},"freeTextAnswersStats":[]}`
		if !bytes.Equal(msgStr, []byte(expected)) {
			t.Errorf("Expected message to be %s, got %s", expected, msgStr)
		}
		if quizGame.currentQuestionIndex != 0 {
			t.Errorf("Expected currentQuestionIndex to be 0, got %d", quizGame.currentQuestionIndex)
		}
	})

	t.Run("Answer first question - same player answers twice", func(t *testing.T) {
		stats, err := quizGame.AnswerMCQuestion(101, []int{1001}, &User{Login: "login1"})
		if stats != nil {
			t.Fatalf("Expected stats to be nil, got %+v", stats)
		}
		if err.Error() != "user login1 already answered question 101" {
			t.Fatalf("Expected error, got %s", err)
		}
		if quizGame.currentQuestionIndex != 0 {
			t.Fatalf("Expected currentQuestionIndex to be 0, got %d", quizGame.currentQuestionIndex)
		}
	})

	t.Run("Answer first question - second player", func(t *testing.T) {
		stats, err := quizGame.AnswerMCQuestion(101, []int{1002}, &User{Login: "login2"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if stats == nil {
			t.Fatal("Expected stats, got nil")
		}
		// generate json for message and compare with expected
		msgStr, err := json.Marshal(stats)
		if err != nil {
			t.Errorf("Error marshaling JsonMessage: %v\n", err)
		}
		expected := `{"type":4,"action":4,"questionId":101,"status":2,"learnersCount":2,"answeredCount":2,"answersStats":{"1001":{"answerId":1001,"count":1,"correct":1},"1002":{"answerId":1002,"count":1,"correct":0}},"freeTextAnswersStats":[]}`
		if !bytes.Equal(msgStr, []byte(expected)) {
			t.Errorf("Expected message to be %s, got %s", expected, msgStr)
		}
		if quizGame.currentQuestionIndex != 0 {
			t.Errorf("Expected currentQuestionIndex to be 0, got %d", quizGame.currentQuestionIndex)
		}
	})

	t.Run("Next question - try to answer but timeout (robustness)", func(t *testing.T) {
		quizGame.NextQuizQuestionMessage()
		quizGame.questionTimer = nil // simulate timeout
		stats, err := quizGame.AnswerMCQuestion(102, []int{1003}, &User{Login: "login1"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if stats == nil {
			t.Fatal("Expected stats, got nil")
		}
		// generate json for message and compare with expected
		msgStr, err := json.Marshal(stats)
		if err != nil {
			t.Errorf("Error marshaling JsonMessage: %v\n", err)
		}
		expected := `{"type":4,"action":4,"questionId":102,"status":1,"learnersCount":2,"answeredCount":0,"answersStats":{},"freeTextAnswersStats":[]}`
		if !bytes.Equal(msgStr, []byte(expected)) {
			t.Errorf("Expected message to be %s, got %s", expected, msgStr)
		}
		if quizGame.currentQuestionIndex != 1 {
			t.Errorf("Expected currentQuestionIndex to be 1, got %d", quizGame.currentQuestionIndex)
		}
	})

	t.Run("Next question  => should be quiz game stats", func(t *testing.T) {
		stats := quizGame.NextQuizQuestionMessage()
		// generate json for message and compare with expected
		msgStr, err := json.Marshal(stats)
		if err != nil {
			t.Errorf("Error marshaling JsonMessage: %v\n", err)
		}
		expected := `{"type":4,"action":6,"quizId":1,"learnersCount":2,"playerStats":{"login1":{"playerLogin":"login1","countAnswered":1,"countCorrect":1},"login2":{"playerLogin":"login2","countAnswered":1,"countCorrect":0}}}`
		if !bytes.Equal(msgStr, []byte(expected)) {
			t.Errorf("Expected message to be %s, got %s", expected, msgStr)
		}
		if quizGame.currentQuestionIndex != 2 {
			t.Errorf("Expected currentQuestionIndex to be 2, got %d", quizGame.currentQuestionIndex)
		}
	})
}
