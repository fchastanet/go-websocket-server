package models

import (
	"fmt"
	"log"
	"time"
)

const DEFAULT_TIMEOUT_SECONDS = 30

func (quizGame *QuizGame) Start(quiz *Quiz, user *User) {
	if quizGame.questionTimer != nil {
		quizGame.questionTimer.Stop()
	}
	quizGame.quiz = quiz
	quizGame.questionStats = make(map[int]QuestionStats)
	quizGame.playerStats = make(map[string]PlayerStat)
	quizGame.StartedAt = time.Now()
	quizGame.StartedBy = user.Login
	quizGame.currentQuestionIndex = -1
}

// contains checks if a slice contains a specific element
func contains(slice []int, element int) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

// GetNextQuestion returns the next question in the quiz
func (q *QuizGame) getNextQuestion() *Question {
	q.currentQuestionIndex++
	if q.currentQuestionIndex < len(q.quiz.Questions) {
		return &q.quiz.Questions[q.currentQuestionIndex]
	}
	return nil
}

func (quizGame *QuizGame) NextQuizQuestionMessage() any {
	previousQuestion := quizGame.currentQuizQuestion
	previousQuestionId := -1
	if previousQuestion != nil {
		previousQuestionId = previousQuestion.ID
	}
	question := quizGame.getNextQuestion()
	if question == nil {
		return &QuizStatsMessage{
			Envelope: &Envelope{
				Type:   MESSAGE_TYPE_QUIZ_MESSAGE,
				Action: QUIZ_MESSAGE_ACTION_STATS,
			},
			LearnersCount: len(quizGame.playerStats),
			QuizId:        quizGame.quiz.ID,
			PlayerStats:   quizGame.playerStats,
		}
	}
	if quizGame.questionTimer != nil {
		log.Printf("Stopping timer for question %d\n", previousQuestionId)
		quizGame.questionTimer.Stop()
	}
	questionClone := question.Clone()
	questionClone.startedAt = time.Now()
	quizGame.currentQuizQuestion = &questionClone

	// create timer
	log.Printf("Starting timer for question %d\n", question.ID)
	quizGame.questionTimer = time.AfterFunc(quizGame.questionTimeout, func() {
		log.Printf("Question %d Timed out after %d.\n", question.ID, quizGame.questionTimeout)
		quizGame.questionTimer = nil
		var quizQuestionStatsMessage *QuizQuestionStatsMessage = nil
		var err error
		if question.QuestionType == QUESTION_TYPE_MCQ {
			quizQuestionStatsMessage, err = quizGame.AnswerMCQuestion(question.ID, []int{}, &User{Login: "system"})
		} else if question.QuestionType == QUESTION_TYPE_FREE_TEXT {
			quizQuestionStatsMessage, err = quizGame.AnswerFreeTextQuestion(question.ID, []string{}, &User{Login: "system"})
		}
		if err != nil {
			log.Printf("Error processing timeout: %v\n", err)
		}
		questionStats := quizGame.GetQuestionStatsOrCreate(question.ID)
		for _, answer := range question.Answers {
			answerStat := questionStats.GetAnswerStatsOrCreate(question.ID, answer.ID)
			answerStat.Correct = answer.Correct
			quizQuestionStatsMessage.AnswersStats[answer.ID] = *answerStat
		}
		if err != nil {
			log.Printf("Error processing timeout: %v\n", err)
		}
		if quizQuestionStatsMessage != nil {
			err := quizGame.commandServices.MessageSender(
				&User{Login: "system", SessionID: quizGame.SessionID},
				quizQuestionStatsMessage,
			)
			if err != nil {
				log.Printf("Error sending timeout message: %v\n", err)
			}
		}
	})

	// remove correct answers from the question
	for i := range questionClone.Answers {
		answer := &questionClone.Answers[i]
		answer.Correct = ANSWER_CORRECT_UNKNOWN
	}

	return &QuizQuestionMessage{
		Envelope: &Envelope{
			Type:   MESSAGE_TYPE_QUIZ_MESSAGE,
			Action: QUIZ_MESSAGE_ACTION_QUESTION_START,
		},
		QuizInfo: QuizInfo{
			ID:        quizGame.quiz.ID,
			Title:     quizGame.quiz.Title,
			Type:      quizGame.quiz.Type,
			URL:       quizGame.quiz.URL,
			StartedBy: quizGame.StartedBy,
		},
		Question:       questionClone,
		QuestionNumber: quizGame.currentQuestionIndex + 1,
		QuestionCount:  len(quizGame.quiz.Questions),
		Timeout:        DEFAULT_TIMEOUT_SECONDS,
	}
}

func (quizGame *QuizGame) GetQuestionStatsOrCreate(questionId int) *QuestionStats {
	stats, ok := quizGame.questionStats[questionId]
	if !ok {
		stats = QuestionStats{
			QuestionID:           questionId,
			AnswersStats:         make(map[int]AnswerStat),
			FreeTextAnswersStats: make([]FreeTextAnswerStat, 0),
			PlayerStats:          make(map[string]QuestionPlayerStat),
		}
		quizGame.questionStats[questionId] = stats
	}
	return &stats
}

func (quizGame *QuestionStats) GetAnswerStatsOrCreate(questionId, answerId int) *AnswerStat {
	stats, ok := quizGame.AnswersStats[answerId]
	if !ok {
		stats = AnswerStat{
			AnswerID: answerId,
			Count:    0,
			Correct:  ANSWER_CORRECT_UNKNOWN,
		}
		quizGame.AnswersStats[answerId] = stats
	}
	return &stats
}

func (quizGame *QuizGame) AnswerMCQuestion(
	questionId int, answers []int, user *User,
) (*QuizQuestionStatsMessage, error) {
	if quizGame.quiz == nil {
		return nil, fmt.Errorf("quiz not started")
	}
	if quizGame.currentQuestionIndex < 0 {
		return nil, fmt.Errorf("quiz not started")
	}
	if quizGame.currentQuestionIndex >= len(quizGame.quiz.Questions) {
		return nil, fmt.Errorf("quiz ended")
	}
	question := quizGame.quiz.Questions[quizGame.currentQuestionIndex]
	if question.ID != questionId {
		return nil, fmt.Errorf("question ID %d does not match current question ID %d", questionId, question.ID)
	}
	if question.QuestionType != QUESTION_TYPE_MCQ {
		return nil, fmt.Errorf("question ID %d is not a multiple choice question", questionId)
	}
	questionStats := quizGame.GetQuestionStatsOrCreate(questionId)
	if quizGame.questionTimer == nil {
		// robustness: in normal cases the server would
		// have already sent a timeout message
		questionStats.QuestionStatus = QUESTION_STATUS_TIMEOUT
		quizGame.questionStats[questionId] = *questionStats
		return quizGame.getQuizQuestionStatsMessage(questionId, QUIZ_MESSAGE_ACTION_QUESTION_END), nil
	}
	questionPlayerStats, ok := questionStats.PlayerStats[user.Login]
	if ok {
		if questionPlayerStats.Correct != ANSWER_CORRECT_UNKNOWN {
			return nil, fmt.Errorf("user %s already answered question %d", user.Login, questionId)
		}
	} else {
		questionPlayerStats = QuestionPlayerStat{
			PlayerLogin: user.Login,
			Correct:     ANSWER_CORRECT_UNKNOWN,
		}
	}

	// check if answers ids could be incorrect
	validAnswerIds := []int{}
	for _, answer := range question.Answers {
		validAnswerIds = append(validAnswerIds, answer.ID)
	}
	for _, answerId := range answers {
		if !contains(validAnswerIds, answerId) {
			return nil, fmt.Errorf("answer ID %d is invalid", answerId)
		}
	}

	// check if answers are correct
	questionAnsweredCorrectly := true
	for _, answer := range question.Answers {
		answerStat := questionStats.GetAnswerStatsOrCreate(questionId, answer.ID)
		answerStat.Correct = answer.Correct

		// check if answerId is in answers
		if contains(answers, answer.ID) {
			answerStat.Count++
			if answer.Correct != ANSWER_CORRECT_CORRECT {
				questionAnsweredCorrectly = false
			}
		} else {
			if answer.Correct != ANSWER_CORRECT_INCORRECT {
				questionAnsweredCorrectly = false
			}
		}
		questionStats.AnswersStats[answer.ID] = *answerStat
	}
	if questionAnsweredCorrectly {
		if questionAnsweredCorrectly {
			questionPlayerStats.Correct = ANSWER_CORRECT_CORRECT
		}
	}
	questionStats.PlayerStats[user.Login] = questionPlayerStats

	// consolidate player stats
	playerStat, ok := quizGame.playerStats[user.Login]
	if !ok {
		playerStat = PlayerStat{
			PlayerLogin:   user.Login,
			CountAnswered: 0,
			CountCorrect:  0,
		}
	}
	playerStat.PlayerLogin = user.Login
	playerStat.CountAnswered++
	if questionAnsweredCorrectly {
		playerStat.CountCorrect += 1
	}
	quizGame.playerStats[user.Login] = playerStat

	// end of quiz if all players have answered
	action := QUIZ_MESSAGE_ACTION_QUESTION_STATS
	playersCount := quizGame.GetConnectedPlayersCount()
	answeredCount := len(questionStats.PlayerStats)
	questionStatus := QUESTION_STATUS_IN_PROGRESS
	if playersCount == answeredCount {
		// reset the timer
		if quizGame.questionTimer != nil {
			log.Printf("Stopping timer for question %d\n", questionId)
			quizGame.questionTimer.Stop()
		}
		// all players have answered
		questionStatus = QUESTION_STATUS_ENDED
		action = QUIZ_MESSAGE_ACTION_QUESTION_END
		quizGame.questionTimer.Stop()
		quizGame.questionTimer = nil
	}
	questionStats.QuestionStatus = questionStatus
	quizGame.questionStats[questionId] = *questionStats

	// message generation
	return quizGame.getQuizQuestionStatsMessage(questionId, action), nil
}

func (quizGame *QuizGame) AnswerFreeTextQuestion(
	questionId int, answers []string, user *User,
) (*QuizQuestionStatsMessage, error) {
	if quizGame.quiz == nil {
		return nil, fmt.Errorf("quiz not started")
	}
	if quizGame.currentQuestionIndex < 0 {
		return nil, fmt.Errorf("quiz not started")
	}
	if quizGame.currentQuestionIndex >= len(quizGame.quiz.Questions) {
		return nil, fmt.Errorf("quiz ended")
	}
	question := quizGame.quiz.Questions[quizGame.currentQuestionIndex]
	if question.ID != questionId {
		return nil, fmt.Errorf("question ID %d does not match current question ID %d", questionId, question.ID)
	}
	if question.QuestionType != QUESTION_TYPE_FREE_TEXT {
		return nil, fmt.Errorf("question ID %d is not a free text question", questionId)
	}

	questionStats := quizGame.GetQuestionStatsOrCreate(questionId)
	if quizGame.questionTimer == nil {
		// robustness: in normal cases the server would
		// have already sent a timeout message
		questionStats.QuestionStatus = QUESTION_STATUS_TIMEOUT
		quizGame.questionStats[questionId] = *questionStats
		return quizGame.getQuizQuestionStatsMessage(questionId, QUIZ_MESSAGE_ACTION_QUESTION_END), nil
	}
	questionPlayerStats, ok := questionStats.PlayerStats[user.Login]
	if ok {
		if questionPlayerStats.Correct != ANSWER_CORRECT_UNKNOWN {
			return nil, fmt.Errorf("user %s already answered question %d", user.Login, questionId)
		}
	} else {
		questionPlayerStats = QuestionPlayerStat{
			PlayerLogin: user.Login,
			Correct:     ANSWER_CORRECT_UNKNOWN,
		}
	}

	// check if answers are correct
	questionPlayerStats.Correct = ANSWER_CORRECT_CORRECT
	questionStats.PlayerStats[user.Login] = questionPlayerStats
	questionStats.FreeTextAnswersStats = append(questionStats.FreeTextAnswersStats, FreeTextAnswerStat{
		Text:  answers[0],
		Login: user.Login,
	})

	// consolidate player stats
	playerStat, ok := quizGame.playerStats[user.Login]
	if !ok {
		playerStat = PlayerStat{
			PlayerLogin:   user.Login,
			CountAnswered: 0,
			CountCorrect:  0,
		}
	}
	playerStat.PlayerLogin = user.Login
	playerStat.CountAnswered++
	quizGame.playerStats[user.Login] = playerStat

	// end of quiz if all players have answered
	action := QUIZ_MESSAGE_ACTION_QUESTION_STATS
	playersCount := quizGame.GetConnectedPlayersCount()
	answeredCount := len(questionStats.PlayerStats)
	questionStatus := QUESTION_STATUS_IN_PROGRESS
	if playersCount == answeredCount {
		// all players have answered
		questionStatus = QUESTION_STATUS_ENDED
		action = QUIZ_MESSAGE_ACTION_QUESTION_END
		quizGame.questionTimer.Stop()
		quizGame.questionTimer = nil
	}
	questionStats.QuestionStatus = questionStatus
	quizGame.questionStats[questionId] = *questionStats

	// message generation
	return quizGame.getQuizQuestionStatsMessage(questionId, action), nil
}

func (quizGame *QuizGame) getQuizQuestionStatsMessage(questionId int, action QuizMessageAction) *QuizQuestionStatsMessage {
	questionStats, ok := quizGame.questionStats[questionId]
	if !ok {
		return nil
	}
	env := &Envelope{
		Type:   MESSAGE_TYPE_QUIZ_MESSAGE,
		Action: action,
	}

	return &QuizQuestionStatsMessage{
		Envelope:             env,
		QuestionID:           questionId,
		Status:               questionStats.QuestionStatus,
		LearnersCount:        quizGame.GetConnectedPlayersCount(),
		AnsweredCount:        len(questionStats.PlayerStats),
		AnswersStats:         questionStats.AnswersStats,
		FreeTextAnswersStats: questionStats.FreeTextAnswersStats,
	}
}
