package event_as_test

type EventTestService interface {
	// GetTest получение теста
	GetTest()
	// GetQuestion получение конкретного вопроса
	GetQuestion()
	// SetAnswer Установить ответ на вопрос
	SetAnswer()
}
