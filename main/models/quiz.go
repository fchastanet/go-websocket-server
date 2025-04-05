package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type AnswerStat struct {
	AnswerID int           `json:"answerId"`
	Count    int           `json:"count"`
	Correct  AnswerCorrect `json:"correct"`
}

type FreeTextAnswerStat struct {
	Text  string `json:"text"`
	Login string `json:"login"`
}

type QuestionPlayerStat struct {
	PlayerLogin string        `json:"playerLogin"`
	Correct     AnswerCorrect `json:"correct"`
}

type QuestionStats struct {
	QuestionID           int                           `json:"questionId"`
	QuestionStatus       QuestionStatus                `json:"status"`
	AnswersStats         map[int]AnswerStat            `json:"answersStats"`
	FreeTextAnswersStats []FreeTextAnswerStat          `json:"freeTextAnswersStats"`
	PlayerStats          map[string]QuestionPlayerStat `json:"playerStats"`
}

type PlayerStat struct {
	PlayerLogin   string `json:"playerLogin"`
	CountAnswered int    `json:"countAnswered"`
	CountCorrect  int    `json:"countCorrect"`
}

type QuizGame struct {
	SessionID                string
	quiz                     *Quiz
	questionStats            map[int]QuestionStats
	playerStats              map[string]PlayerStat
	currentQuestionIndex     int
	StartedBy                string
	StartedAt                time.Time
	players                  []string
	GetConnectedPlayersCount func() int

	// timeout management
	currentQuizQuestion *Question
	questionTimer       *time.Timer
	questionTimeout     time.Duration
	commandServices     CommandServices
}

// Quiz represents a complete quiz with questions
type Quiz struct {
	ID        int        `json:"id"`
	Type      QuizType   `json:"type"`
	Title     string     `json:"title"`
	URL       string     `json:"url"`
	Questions []Question `json:"questions"`
}

// Question represents a single quiz question
type Question struct {
	ID           int          `json:"id"`
	Question     string       `json:"question"`
	QuestionType QuestionType `json:"questionType"`
	URL          string       `json:"url"`
	Answers      []Answer     `json:"answers"`
	startedAt    time.Time
}

type AnswerCorrect int

const (
	ANSWER_CORRECT_UNKNOWN   AnswerCorrect = -1
	ANSWER_CORRECT_INCORRECT AnswerCorrect = 0
	ANSWER_CORRECT_CORRECT   AnswerCorrect = 1
)

// Answer represents a possible answer to a question
type Answer struct {
	ID      int           `json:"id"`
	Title   string        `json:"title"`
	URL     string        `json:"url"`
	Correct AnswerCorrect `json:"correct,omitempty"`
}

// ParseQuiz parses a JSON string into a Quiz struct
func ParseQuiz(jsonData string) (*Quiz, error) {
	var quiz Quiz
	err := json.Unmarshal([]byte(jsonData), &quiz)
	if err != nil {
		return nil, fmt.Errorf("error parsing quiz JSON: %v", err)
	}
	return &quiz, nil
}

func (q *Quiz) Clone() *Quiz {
	clone := *q
	clone.Questions = make([]Question, len(q.Questions))
	for i, question := range q.Questions {
		clone.Questions[i] = question.Clone()
	}
	return &clone
}

func (q *Question) Clone() Question {
	clone := *q
	clone.Answers = make([]Answer, len(q.Answers))
	for i, answer := range q.Answers {
		clone.Answers[i] = answer.Clone()
	}
	return clone
}

func (a *Answer) Clone() Answer {
	return *a
}

// GetAnswerByID returns an answer by its ID
func (q *Question) GetAnswerByID(id int) *Answer {
	for _, answer := range q.Answers {
		if answer.ID == id {
			return &answer
		}
	}
	return nil
}

// GetQuestionByID returns a question by its ID
func (q *Quiz) GetQuestionByID(id int) *Question {
	for _, question := range q.Questions {
		if question.ID == id {
			return &question
		}
	}
	return nil
}
