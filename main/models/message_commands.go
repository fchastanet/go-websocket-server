package models

import (
	"fmt"
	"time"
)

type Command interface {
	Execute(user *User, session *Session, commandServices CommandServices) error
}

type CommandServices struct {
	MessageSender                              func(user *User, message interface{}) error
	SendUserConnectMessageForAllUsersInSession func(session *Session) error
	GetQuiz                                    func(quizId int) (quiz *Quiz, err error)
}

func (msg *UserConnectMessage) Execute(
	user *User, session *Session, commandServices CommandServices,
) error {
	user.Login = msg.From.Id
	return commandServices.SendUserConnectMessageForAllUsersInSession(session)
}

func (msg *UserDisconnectMessage) Execute(
	user *User, session *Session, commandServices CommandServices,
) error {
	return nil
}

func (msg *Message) Execute(
	user *User, session *Session, commandServices CommandServices,
) error {
	msg.From = Recipient{
		Type: RECIPIENT_TYPE_LEARNER,
		Id:   user.Login,
	}
	return commandServices.MessageSender(user, msg)
}

func nextQuestion(
	user *User, session *Session, commandServices CommandServices,
) error {
	quizMsg := session.QuizGame.NextQuizQuestionMessage()
	if quizMsg == nil {
		return nil
	}
	return commandServices.MessageSender(user, quizMsg)
}

func startQuiz(session *Session, commandServices CommandServices, quiz *Quiz, user *User) {
	session.QuizGame.questionTimeout = DEFAULT_TIMEOUT_SECONDS * time.Second
	session.QuizGame.commandServices = commandServices
	session.QuizGame.Start(quiz, user)
}

func (msg *QuizStartMessage) Execute(
	user *User, session *Session, commandServices CommandServices,
) error {
	quizId := int(msg.QuizId)
	quiz, err := commandServices.GetQuiz(quizId)
	if err != nil {
		return fmt.Errorf("error getting quiz with id: %d", quizId)
	}
	startQuiz(session, commandServices, quiz, user)

	return nextQuestion(user, session, commandServices)
}

func (msg *QuizNextQuestionMessage) Execute(
	user *User, session *Session, commandServices CommandServices,
) error {
	quizId := int(msg.QuizId)
	quiz, err := commandServices.GetQuiz(quizId)
	if err != nil {
		return fmt.Errorf("error getting quiz with id: %d", quizId)
	}
	if session.QuizGame.quiz == nil {
		startQuiz(session, commandServices, quiz, user)
	}

	return nextQuestion(user, session, commandServices)
}

func (msg *QuizQuestionMessage) Execute(
	user *User, session *Session, commandServices CommandServices,
) error {
	return fmt.Errorf("QuizQuestionMessage can only be sent by the server")
}

func (msg *QuizLearnerAnswerMessage) Execute(
	user *User, session *Session, commandServices CommandServices,
) error {
	quizQuestionStatsMessage, err := session.QuizGame.AnswerMCQuestion(msg.QuestionId, msg.Answers, user)
	if err != nil {
		return fmt.Errorf("error processing learner answer: %v", err)
	}
	if quizQuestionStatsMessage != nil {
		err := commandServices.MessageSender(user, quizQuestionStatsMessage)
		if err != nil {
			return fmt.Errorf("error sending quiz question stats message: %v", err)
		}
	}
	return nil
}

func (msg *QuizLearnerAnswerFreeTextMessage) Execute(
	user *User, session *Session, commandServices CommandServices,
) error {
	quizQuestionStatsMessage, err := session.QuizGame.AnswerFreeTextQuestion(msg.QuestionId, msg.Answers, user)
	if err != nil {
		return fmt.Errorf("error processing learner answer: %v", err)
	}
	if quizQuestionStatsMessage != nil {
		err := commandServices.MessageSender(user, quizQuestionStatsMessage)
		if err != nil {
			return fmt.Errorf("error sending quiz question stats message: %v", err)
		}
	}
	return nil
}

func (msg *QuizQuestionStatsMessage) Execute(
	user *User, session *Session, commandServices CommandServices,
) error {
	return fmt.Errorf("QuizQuestionStatsMessage can only be sent by the server")
}

func (msg *QuizQuestionEndMessage) Execute(
	user *User, session *Session, commandServices CommandServices,
) error {
	return fmt.Errorf("QuizQuestionEndMessage can only be sent by the server")
}
