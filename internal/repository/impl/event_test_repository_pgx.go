package impl

import (
	"context"
	"ne_noy/internal/model/events/as_test"
	"ne_noy/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type eventTestRepositoryPgx struct {
	pool *pgxpool.Pool
}

func NewEventTestRepository(pool *pgxpool.Pool) repository.EventTestRepository {
	return &eventTestRepositoryPgx{pool: pool}
}

func (e *eventTestRepositoryPgx) GetTest(ctx context.Context, testID uuid.UUID) (*as_test.AsTest, error) {
	test, err := e.getTestProfile(ctx, testID)
	if err != nil {
		return nil, err
	}

	questions, err := e.getQuestionsByTest(ctx, testID)
	if err != nil {
		return nil, err
	}
	test.Questions = questions

	return test, nil
}

func (e *eventTestRepositoryPgx) GetQuestion(ctx context.Context, testID, questionID uuid.UUID) (*as_test.Question, error) {
	row := e.pool.QueryRow(ctx, `
		SELECT id, created_at, updated_at, text, type, event_id, q_order
		FROM questions
		WHERE id = $1 AND event_id = $2
	`, questionID, testID)

	question, err := scanQuestion(row)
	if err != nil {
		return nil, err
	}

	question.Answers, err = e.getAnswersByQuestion(ctx, question.ID)
	if err != nil {
		return nil, err
	}

	return question, nil
}

func (e *eventTestRepositoryPgx) CreateTest(ctx context.Context, test *as_test.AsTest) (*as_test.AsTest, error) {
	testID := uuid.New()

	// Все поля профиля теста создаются одной вставкой, чтобы вернуть стабильный ID для следующих операций с вопросами.
	_, err := e.pool.Exec(ctx, `
		INSERT INTO event_as_tests (id, name, description, cover, status, starts_at, ends_at, ext_link_id, attempts, vk_post_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, testID, test.Name, test.Description, test.Cover, test.Status, test.StartsAt, test.EndsAt, test.ExtLinkID, test.Attempts, test.VkPostID)
	if err != nil {
		return nil, err
	}

	return e.GetTest(ctx, testID)
}

func (e *eventTestRepositoryPgx) UpdateTest(ctx context.Context, testID uuid.UUID, update as_test.AsTest) (*as_test.AsTest, error) {
	// COALESCE оставляет старое значение, если сервис не передал новое поле в модели обновления.
	commandTag, err := e.pool.Exec(ctx, `
		UPDATE event_as_tests
		SET
			name = COALESCE($2, name),
			description = COALESCE($3, description),
			cover = COALESCE($4, cover),
			status = COALESCE($5, status),
			starts_at = COALESCE($6, starts_at),
			ends_at = COALESCE($7, ends_at),
			ext_link_id = COALESCE($8, ext_link_id),
			attempts = COALESCE($9, attempts),
			vk_post_id = COALESCE($10, vk_post_id),
			updated_at = now()
		WHERE id = $1
	`, testID, nullableString(update.Name), update.Description, update.Cover, nullableString(update.Status), nullableTime(update.StartsAt), update.EndsAt, update.ExtLinkID, nullableInt(update.Attempts), update.VkPostID)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() == 0 {
		return nil, pgx.ErrNoRows
	}

	return e.GetTest(ctx, testID)
}

func (e *eventTestRepositoryPgx) DeleteTest(ctx context.Context, testID uuid.UUID) error {
	commandTag, err := e.pool.Exec(ctx, `
		DELETE FROM event_as_tests
		WHERE id = $1
	`, testID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	// Вопросы, варианты ответов и пользовательские ответы удаляются каскадно через внешние ключи.
	return nil
}

func (e *eventTestRepositoryPgx) AddQuestion(ctx context.Context, testID uuid.UUID, question as_test.Question) (*as_test.Question, error) {
	questionID := uuid.New()

	// Вопрос связывается с event_as_tests через event_id; порядок хранится явно для стабильной выдачи теста.
	_, err := e.pool.Exec(ctx, `
		INSERT INTO questions (id, text, type, event_id, q_order)
		VALUES ($1, $2, $3, $4, $5)
	`, questionID, question.Text, question.Type, testID, question.QOrder)
	if err != nil {
		return nil, err
	}

	return e.GetQuestion(ctx, testID, questionID)
}

func (e *eventTestRepositoryPgx) AddAnswer(ctx context.Context, questionID uuid.UUID, answer as_test.Answer) (*as_test.Answer, error) {
	answerID := uuid.New()

	// Баллы храним у варианта ответа, чтобы при пользовательском выборе быстро посчитать результат.
	_, err := e.pool.Exec(ctx, `
		INSERT INTO answers (id, question_id, is_correct, text, points)
		VALUES ($1, $2, $3, $4, $5)
	`, answerID, questionID, answer.IsCorrect, answer.Text, answer.Points)
	if err != nil {
		return nil, err
	}

	return e.getAnswerByID(ctx, answerID)
}

func (e *eventTestRepositoryPgx) SetUserAnswer(ctx context.Context, userAnswer as_test.UserAnswer) (*as_test.UserAnswer, error) {
	points, err := e.resolveUserAnswerPoints(ctx, userAnswer.QuestionID, userAnswer.AnswerID)
	if err != nil {
		return nil, err
	}

	userAnswerID := uuid.New()
	row := e.pool.QueryRow(ctx, `
		INSERT INTO user_answers (id, user_id, question_id, answer_id, text, points)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at, user_id, question_id, answer_id, text, points
	`, userAnswerID, userAnswer.UserID, userAnswer.QuestionID, userAnswer.AnswerID, userAnswer.Text, points)

	var saved as_test.UserAnswer
	if err = row.Scan(
		&saved.ID, &saved.CreatedAt, &saved.UpdatedAt, &saved.UserID, &saved.QuestionID, &saved.AnswerID, &saved.Text, &saved.Points,
	); err != nil {
		return nil, err
	}

	return &saved, nil
}

func (e *eventTestRepositoryPgx) getTestProfile(ctx context.Context, testID uuid.UUID) (*as_test.AsTest, error) {
	row := e.pool.QueryRow(ctx, `
		SELECT id, created_at, updated_at, name, description, cover, status, starts_at, ends_at, ext_link_id, attempts, vk_post_id
		FROM event_as_tests
		WHERE id = $1
	`, testID)

	var test as_test.AsTest
	if err := row.Scan(
		&test.ID, &test.CreatedAt, &test.UpdatedAt, &test.Name, &test.Description, &test.Cover, &test.Status,
		&test.StartsAt, &test.EndsAt, &test.ExtLinkID, &test.Attempts, &test.VkPostID,
	); err != nil {
		return nil, err
	}

	return &test, nil
}

func (e *eventTestRepositoryPgx) getQuestionsByTest(ctx context.Context, testID uuid.UUID) ([]*as_test.Question, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT id, created_at, updated_at, text, type, event_id, q_order
		FROM questions
		WHERE event_id = $1
		ORDER BY q_order ASC, created_at ASC
	`, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions := make([]*as_test.Question, 0)
	for rows.Next() {
		question, scanErr := scanQuestion(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		question.Answers, scanErr = e.getAnswersByQuestion(ctx, question.ID)
		if scanErr != nil {
			return nil, scanErr
		}
		questions = append(questions, question)
	}

	return questions, rows.Err()
}

func (e *eventTestRepositoryPgx) getAnswersByQuestion(ctx context.Context, questionID uuid.UUID) ([]*as_test.Answer, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT id, created_at, updated_at, question_id, is_correct, text, points
		FROM answers
		WHERE question_id = $1
		ORDER BY created_at ASC
	`, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	answers := make([]*as_test.Answer, 0)
	for rows.Next() {
		answer, scanErr := scanAnswer(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		answers = append(answers, answer)
	}

	return answers, rows.Err()
}

func (e *eventTestRepositoryPgx) getAnswerByID(ctx context.Context, answerID uuid.UUID) (*as_test.Answer, error) {
	row := e.pool.QueryRow(ctx, `
		SELECT id, created_at, updated_at, question_id, is_correct, text, points
		FROM answers
		WHERE id = $1
	`, answerID)

	return scanAnswer(row)
}

func (e *eventTestRepositoryPgx) resolveUserAnswerPoints(ctx context.Context, questionID uuid.UUID, answerID *uuid.UUID) (int, error) {
	if answerID == nil {
		// Для открытого ответа баллы не начисляются автоматически, но проверяем существование вопроса.
		var exists bool
		err := e.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM questions WHERE id = $1)`, questionID).Scan(&exists)
		if err != nil {
			return 0, err
		}
		if !exists {
			return 0, pgx.ErrNoRows
		}
		return 0, nil
	}

	var points int
	err := e.pool.QueryRow(ctx, `
		SELECT points
		FROM answers
		WHERE id = $1 AND question_id = $2
	`, *answerID, questionID).Scan(&points)
	if err != nil {
		return 0, err
	}

	return points, nil
}

func (e *eventTestRepositoryPgx) GetUserAnswersByEvent(ctx context.Context, eventID, userID uuid.UUID) ([]as_test.UserAnswer, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT ua.id, ua.created_at, ua.updated_at, ua.user_id, ua.question_id, ua.answer_id, ua.text, ua.points
		FROM user_answers ua
		JOIN questions q ON ua.question_id = q.id
		WHERE q.event_id = $1 AND ua.user_id = $2
		ORDER BY q.q_order ASC
	`, eventID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []as_test.UserAnswer
	for rows.Next() {
		var ua as_test.UserAnswer
		if err := rows.Scan(
			&ua.ID, &ua.CreatedAt, &ua.UpdatedAt, &ua.UserID, &ua.QuestionID, &ua.AnswerID, &ua.Text, &ua.Points,
		); err != nil {
			return nil, err
		}
		result = append(result, ua)
	}
	return result, rows.Err()
}

func (e *eventTestRepositoryPgx) GetAllUserAnswersByEvent(ctx context.Context, eventID uuid.UUID) ([]as_test.UserAnswer, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT
			ua.id, ua.created_at, ua.updated_at,
			ua.user_id, ua.question_id, ua.answer_id, ua.text, ua.points,
			u.id, u.vk_id, u.first_name, u.last_name, u.photo_url,
			a.is_correct
		FROM user_answers ua
		JOIN questions q ON ua.question_id = q.id
		JOIN users u ON ua.user_id = u.id
		LEFT JOIN answers a ON ua.answer_id = a.id
		WHERE q.event_id = $1
		ORDER BY ua.user_id, ua.created_at
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []as_test.UserAnswer
	for rows.Next() {
		var ua as_test.UserAnswer
		var isCorrect *bool
		if err := rows.Scan(
			&ua.ID, &ua.CreatedAt, &ua.UpdatedAt,
			&ua.UserID, &ua.QuestionID, &ua.AnswerID, &ua.Text, &ua.Points,
			&ua.User.ID, &ua.User.VkID, &ua.User.FirstName, &ua.User.LastName, &ua.User.PhotoURL,
			&isCorrect,
		); err != nil {
			return nil, err
		}
		if isCorrect != nil {
			ua.Answer = &as_test.Answer{IsCorrect: *isCorrect}
		}
		result = append(result, ua)
	}
	return result, rows.Err()
}

func scanQuestion(row pgx.Row) (*as_test.Question, error) {
	var question as_test.Question
	if err := row.Scan(
		&question.ID, &question.CreatedAt, &question.UpdatedAt, &question.Text, &question.Type, &question.EventID, &question.QOrder,
	); err != nil {
		return nil, err
	}

	return &question, nil
}

func scanAnswer(row pgx.Row) (*as_test.Answer, error) {
	var answer as_test.Answer
	if err := row.Scan(
		&answer.ID, &answer.CreatedAt, &answer.UpdatedAt, &answer.QuestionID, &answer.IsCorrect, &answer.Text, &answer.Points,
	); err != nil {
		return nil, err
	}

	return &answer, nil
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func nullableInt(value int) *int {
	if value == 0 {
		return nil
	}
	return &value
}

func nullableTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}
