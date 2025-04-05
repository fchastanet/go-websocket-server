package models

type MessageType int

const (
	MESSAGE_TYPE_USER_CONNECTED    MessageType = 0
	MESSAGE_TYPE_USER_DISCONNECTED MessageType = 1
	MESSAGE_TYPE_MESSAGE           MessageType = 2
	MESSAGE_TYPE_NOTIFICATION      MessageType = 3
	MESSAGE_TYPE_QUIZ_MESSAGE      MessageType = 4
)

// RecipientType represents the type of entity (session or login)
type RecipientType int

const (
	RECIPIENT_TYPE_SESSION RecipientType = 0
	RECIPIENT_TYPE_LEARNER RecipientType = 1
)

// Recipient represents a user or session with type and ID
type Recipient struct {
	Type RecipientType `json:"type"`
	Id   string        `json:"id"`
}

// JsonMessage represents a JSON message with sender, recipients and content
type Envelope struct {
	Type      MessageType       `json:"type"`
	Action    QuizMessageAction `json:"action,omitempty"` // QuizMessageAction
	Timestamp int               `json:"timestamp,omitempty"`
	ClientId  string            `json:"clientId,omitempty"`
}

type UserConnectMessage struct {
	From Recipient   `json:"from"`
	To   []Recipient `json:"to"`
	*Envelope
}

type UserDisconnectMessage struct {
	From Recipient   `json:"from"`
	To   []Recipient `json:"to"`
	*Envelope
}

type Message struct {
	From Recipient   `json:"from"`
	To   []Recipient `json:"to"`
	Msg  string      `json:"msg"`
	*Envelope
}

// QuizMessageAction defines the possible actions for quiz messages
type QuizMessageAction int

const (
	QUIZ_MESSAGE_ACTION_START QuizMessageAction = iota
	QUIZ_MESSAGE_ACTION_QUESTION_START
	QUIZ_MESSAGE_ACTION_LEARNER_ANSWER
	QUIZ_MESSAGE_ACTION_QUESTION_STATS
	QUIZ_MESSAGE_ACTION_QUESTION_END
	QUIZ_MESSAGE_ACTION_NEXT_QUESTION
	QUIZ_MESSAGE_ACTION_STATS
	QUIZ_MESSAGE_ACTION_LEARNER_ANSWER_FREE_TEXT
)

// QuizQuestionStatus defines the possible statuses of a quiz question
type QuizQuestionStatus int

const (
	QUIZ_QUESTION_STATUS_IN_PROGRESS QuizQuestionStatus = iota
	QUIZ_QUESTION_STATUS_TIMEOUT
	QUIZ_QUESTION_STATUS_ENDED
)

type QuizStartMessage struct {
	QuizId int `json:"quizId"`
	*Envelope
}

type QuizNextQuestionMessage struct {
	QuizId int `json:"quizId"`
	*Envelope
}

// QuizAnswer represents a single answer option in a quiz question
type QuizAnswer struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type QuizType int

const (
	QUIZ_TYPE_MCQ       QuizType = 0
	QUIZ_TYPE_FREE_TEXT QuizType = 1
)

type QuizInfo struct {
	ID        int      `json:"id"`
	Title     string   `json:"title"`
	Type      QuizType `json:"type"`
	URL       string   `json:"url"`
	StartedBy string   `json:"startedBy"`
}

type QuizQuestionMessage struct {
	*Envelope
	QuizInfo       QuizInfo     `json:"quizInfo"`
	Question       Question     `json:"question"`
	QuestionType   QuestionType `json:"questionType"`
	QuestionNumber int          `json:"questionNumber"`
	QuestionCount  int          `json:"questionCount"`
	Timeout        int          `json:"timeout"`
}

type QuestionStatus int

const (
	QUESTION_STATUS_IN_PROGRESS QuestionStatus = iota
	QUESTION_STATUS_TIMEOUT
	QUESTION_STATUS_ENDED
)

type QuestionType int

const (
	QUESTION_TYPE_MCQ       QuestionType = 0
	QUESTION_TYPE_FREE_TEXT QuestionType = 1
)

type QuizQuestionStatsMessage struct {
	*Envelope
	QuestionID           int                  `json:"questionId"`
	Status               QuestionStatus       `json:"status"`
	LearnersCount        int                  `json:"learnersCount"`
	AnsweredCount        int                  `json:"answeredCount"`
	AnswersStats         map[int]AnswerStat   `json:"answersStats"`
	FreeTextAnswersStats []FreeTextAnswerStat `json:"freeTextAnswersStats"`
}

type QuizQuestionEndMessage struct {
	*QuizQuestionStatsMessage
}

type QuizLearnerAnswerMessage struct {
	*Envelope
	QuestionId int   `json:"questionId"`
	Answers    []int `json:"answers"`
}

type QuizLearnerAnswerFreeTextMessage struct {
	*Envelope
	QuestionId int      `json:"questionId"`
	Answers    []string `json:"answers"`
}

type QuizStatsMessage struct {
	*Envelope
	QuizId        int                   `json:"quizId"`
	LearnersCount int                   `json:"learnersCount"`
	PlayerStats   map[string]PlayerStat `json:"playerStats"`
}
