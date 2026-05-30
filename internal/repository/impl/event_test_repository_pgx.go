package impl

import (
	"context"
	"ne_noy/internal/model"
	"ne_noy/internal/model/events"
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

	attachments, err := e.getEventAttachments(ctx, testID)
	if err != nil {
		return nil, err
	}
	test.Attachments = attachments

	roles, err := e.getEventRoles(ctx, testID)
	if err != nil {
		return nil, err
	}
	test.AvailableRoleCodes = roles

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

	if err = e.replaceEventAttachments(ctx, testID, test.Attachments); err != nil {
		return nil, err
	}

	if len(test.AvailableRoleCodes) > 0 {
		if err = e.replaceEventRoles(ctx, testID, test.AvailableRoleCodes); err != nil {
			return nil, err
		}
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

	if update.Attachments != nil {
		if err := e.replaceEventAttachments(ctx, testID, update.Attachments); err != nil {
			return nil, err
		}
	}

	if update.AvailableRoleCodes != nil {
		if err := e.replaceEventRoles(ctx, testID, update.AvailableRoleCodes); err != nil {
			return nil, err
		}
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

	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err = lockQuestions(ctx, tx, testID); err != nil {
		return nil, err
	}

	if _, err = tx.Exec(ctx, `
		INSERT INTO questions (id, text, type, event_id, q_order)
		VALUES ($1, $2, $3, $4, $5)
	`, questionID, question.Text, question.Type, testID, question.QOrder); err != nil {
		return nil, err
	}

	if err = renumberQuestions(ctx, tx, testID); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return e.GetQuestion(ctx, testID, questionID)
}

func (e *eventTestRepositoryPgx) UpdateQuestion(ctx context.Context, testID, questionID uuid.UUID, update as_test.Question) (*as_test.Question, error) {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err = lockQuestions(ctx, tx, testID); err != nil {
		return nil, err
	}

	commandTag, err := tx.Exec(ctx, `
		UPDATE questions
		SET text       = COALESCE($3, text),
		    type       = COALESCE($4, type),
		    q_order    = COALESCE($5, q_order),
		    updated_at = now()
		WHERE id = $1 AND event_id = $2
	`, questionID, testID, nullableString(update.Text), nullableString(update.Type), nullableInt(update.QOrder))
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() == 0 {
		return nil, pgx.ErrNoRows
	}

	if err = renumberQuestions(ctx, tx, testID); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return e.GetQuestion(ctx, testID, questionID)
}

// lockQuestions захватывает advisory lock на уровне транзакции для данного теста.
// Это гарантирует, что параллельные операции с вопросами одного теста выполняются
// последовательно и не приводят к дедлоку при перенумерации.
func lockQuestions(ctx context.Context, tx pgx.Tx, testID uuid.UUID) error {
	_, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1::text))`, testID)
	return err
}

// renumberQuestions перенумерует все вопросы теста от 1, сохраняя их относительный порядок.
func renumberQuestions(ctx context.Context, tx pgx.Tx, testID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE questions AS q
		SET q_order = sub.rn
		FROM (
			SELECT id, ROW_NUMBER() OVER (ORDER BY q_order, created_at) AS rn
			FROM questions
			WHERE event_id = $1
		) AS sub
		WHERE q.id = sub.id
	`, testID)
	return err
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
		INSERT INTO user_answers (id, user_id, question_id, answer_id, text, points, attempt)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at, user_id, question_id, answer_id, text, points, attempt
	`, userAnswerID, userAnswer.UserID, userAnswer.QuestionID, userAnswer.AnswerID, userAnswer.Text, points, userAnswer.AttemptID)

	var saved as_test.UserAnswer
	if err = row.Scan(
		&saved.ID, &saved.CreatedAt, &saved.UpdatedAt, &saved.UserID, &saved.QuestionID, &saved.AnswerID, &saved.Text, &saved.Points, &saved.AttemptID,
	); err != nil {
		return nil, err
	}

	return &saved, nil
}

func (e *eventTestRepositoryPgx) UpdateUserAnswer(ctx context.Context, userAnswer as_test.UserAnswer) (*as_test.UserAnswer, error) {
	points, err := e.resolveUserAnswerPoints(ctx, userAnswer.QuestionID, userAnswer.AnswerID)
	if err != nil {
		return nil, err
	}

	row := e.pool.QueryRow(ctx, `
		UPDATE user_answers
		SET answer_id  = $3,
		    text       = $4,
		    points     = $5,
		    updated_at = now()
		WHERE user_id = $1 AND question_id = $2
		  AND attempt IS NOT DISTINCT FROM $6
		RETURNING id, created_at, updated_at, user_id, question_id, answer_id, text, points, attempt
	`, userAnswer.UserID, userAnswer.QuestionID, userAnswer.AnswerID, userAnswer.Text, points, userAnswer.AttemptID)

	var saved as_test.UserAnswer
	if err = row.Scan(
		&saved.ID, &saved.CreatedAt, &saved.UpdatedAt,
		&saved.UserID, &saved.QuestionID, &saved.AnswerID, &saved.Text, &saved.Points, &saved.AttemptID,
	); err != nil {
		return nil, err
	}
	return &saved, nil
}

func (e *eventTestRepositoryPgx) CreateAttempt(ctx context.Context, userID, testID uuid.UUID) (*as_test.UserAttempt, error) {
	attemptID := uuid.New()
	row := e.pool.QueryRow(ctx, `
		INSERT INTO user_attempts (id, userid, testid)
		VALUES ($1, $2, $3)
		RETURNING id, userid, testid, started
	`, attemptID, userID, testID)

	var attempt as_test.UserAttempt
	if err := row.Scan(&attempt.ID, &attempt.UserID, &attempt.TestID, &attempt.Started); err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (e *eventTestRepositoryPgx) GetUserAttempts(ctx context.Context, userID, testID uuid.UUID) ([]as_test.UserAttemptInfo, error) {
	rows, err := e.pool.Query(ctx, `
		WITH attempt_points AS (
			SELECT uana.attempt, COALESCE(SUM(uana.points), 0) AS total_points
			FROM user_answers uana
			WHERE uana.attempt IS NOT NULL
			GROUP BY uana.attempt
		),
		ranked AS (
			SELECT
				ua.id,
				ua.userid,
				ua.testid,
				ua.started,
				COALESCE(ap.total_points, 0) AS points,
				ROW_NUMBER() OVER (PARTITION BY ua.userid, ua.testid ORDER BY ua.started) AS attempt_number,
				ROW_NUMBER() OVER (PARTITION BY ua.testid ORDER BY ua.started)            AS order_number
			FROM user_attempts ua
			LEFT JOIN attempt_points ap ON ap.attempt = ua.id
			WHERE ua.testid = $2
		)
		SELECT id, userid, testid, started, points, attempt_number, order_number
		FROM ranked
		WHERE userid = $1
		ORDER BY started
	`, userID, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []as_test.UserAttemptInfo
	for rows.Next() {
		var info as_test.UserAttemptInfo
		if err := rows.Scan(
			&info.ID, &info.UserID, &info.TestID, &info.Started,
			&info.Points, &info.AttemptNumber, &info.OrderNumber,
		); err != nil {
			return nil, err
		}
		result = append(result, info)
	}
	return result, rows.Err()
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

func (e *eventTestRepositoryPgx) GetUserAnswersByEvent(ctx context.Context, eventID, userID uuid.UUID, attemptID *uuid.UUID) ([]as_test.UserAnswer, error) {
	query := `
		SELECT ua.id, ua.created_at, ua.updated_at, ua.user_id, ua.question_id, ua.answer_id, ua.text, ua.points, ua.attempt
		FROM user_answers ua
		JOIN questions q ON ua.question_id = q.id
		WHERE q.event_id = $1 AND ua.user_id = $2
	`
	args := []any{eventID, userID}
	if attemptID != nil {
		query += ` AND ua.attempt = $3`
		args = append(args, *attemptID)
	}
	query += ` ORDER BY q.q_order ASC`

	rows, err := e.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []as_test.UserAnswer
	for rows.Next() {
		var ua as_test.UserAnswer
		if err := rows.Scan(
			&ua.ID, &ua.CreatedAt, &ua.UpdatedAt, &ua.UserID, &ua.QuestionID, &ua.AnswerID, &ua.Text, &ua.Points, &ua.AttemptID,
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

func (e *eventTestRepositoryPgx) GetTestUserAttempts(ctx context.Context, testID uuid.UUID) ([]as_test.UserAttemptWithSelections, error) {
	rows, err := e.pool.Query(ctx, `
		WITH attempt_points AS (
			SELECT uana.attempt, COALESCE(SUM(uana.points), 0) AS total_points
			FROM user_answers uana
			WHERE uana.attempt IS NOT NULL
			GROUP BY uana.attempt
		)
		SELECT
			ua.id,
			ua.userid,
			ua.testid,
			ua.started,
			COALESCE(ap.total_points, 0)                                                      AS points,
			ROW_NUMBER() OVER (PARTITION BY ua.userid, ua.testid ORDER BY ua.started)         AS attempt_number,
			ROW_NUMBER() OVER (PARTITION BY ua.testid ORDER BY ua.started)                    AS order_number,
			u.id, u.vk_id, u.first_name, u.last_name, u.photo_url
		FROM user_attempts ua
		LEFT JOIN attempt_points ap ON ap.attempt = ua.id
		JOIN users u ON u.id = ua.userid
		WHERE ua.testid = $1
		ORDER BY ua.userid, ua.started
	`, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attempts []as_test.UserAttemptWithSelections
	for rows.Next() {
		var a as_test.UserAttemptWithSelections
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.TestID, &a.Started,
			&a.Points, &a.AttemptNumber, &a.OrderNumber,
			&a.User.ID, &a.User.VkID, &a.User.FirstName, &a.User.LastName, &a.User.PhotoURL,
		); err != nil {
			return nil, err
		}
		attempts = append(attempts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(attempts) == 0 {
		return attempts, nil
	}

	// Загружаем выбранные ответы для всех попыток одним запросом.
	selRows, err := e.pool.Query(ctx, `
		SELECT uana.attempt, uana.answer_id
		FROM user_answers uana
		JOIN questions q ON uana.question_id = q.id
		WHERE q.event_id = $1
		  AND uana.attempt IS NOT NULL
		  AND uana.answer_id IS NOT NULL
	`, testID)
	if err != nil {
		return nil, err
	}
	defer selRows.Close()

	selected := make(map[uuid.UUID][]uuid.UUID)
	for selRows.Next() {
		var attemptID, answerID uuid.UUID
		if err := selRows.Scan(&attemptID, &answerID); err != nil {
			return nil, err
		}
		selected[attemptID] = append(selected[attemptID], answerID)
	}
	if err := selRows.Err(); err != nil {
		return nil, err
	}

	for i := range attempts {
		if ids, ok := selected[attempts[i].ID]; ok {
			attempts[i].SelectedAnswers = ids
		} else {
			attempts[i].SelectedAnswers = []uuid.UUID{}
		}
	}

	return attempts, nil
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

func (e *eventTestRepositoryPgx) SetEventOrganizers(ctx context.Context, eventID uuid.UUID, userIDs []uuid.UUID) error {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, `DELETE FROM event_orgs WHERE event_id = $1 AND event_type = 'test'`, eventID); err != nil {
		return err
	}

	for _, userID := range userIDs {
		if _, err = tx.Exec(ctx, `
			INSERT INTO event_orgs (event_id, event_type, user_id) VALUES ($1, 'test', $2)
			ON CONFLICT DO NOTHING
		`, eventID, userID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (e *eventTestRepositoryPgx) GetEventOrganizers(ctx context.Context, eventID uuid.UUID) ([]model.User, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT u.id, u.vk_id, u.first_name, u.last_name, u.photo_url
		FROM users u
		JOIN event_orgs eo ON eo.user_id = u.id
		WHERE eo.event_id = $1 AND eo.event_type = 'test'
		ORDER BY u.first_name ASC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.VkID, &u.FirstName, &u.LastName, &u.PhotoURL); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (e *eventTestRepositoryPgx) getEventAttachments(ctx context.Context, testID uuid.UUID) ([]events.EventAttachment, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT ea.attachment_id, a.url, a.filename
		FROM event_attachments ea
		JOIN attachments a ON a.id = ea.attachment_id
		WHERE ea.event_id = $1 AND ea.event_type = $2::event_type_enum
		ORDER BY ea.created_at ASC
	`, testID, events.EventAsTest)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]events.EventAttachment, 0)
	for rows.Next() {
		var ea events.EventAttachment
		ea.EventID = testID
		ea.EventType = events.EventAsTest
		ea.Attachment = &model.Attachment{}
		if err := rows.Scan(&ea.AttachmentID, &ea.Attachment.Url, &ea.Attachment.Filename); err != nil {
			return nil, err
		}
		if ea.AttachmentID != nil {
			ea.Attachment.ID = *ea.AttachmentID
		}
		result = append(result, ea)
	}
	return result, rows.Err()
}

func (e *eventTestRepositoryPgx) replaceEventAttachments(ctx context.Context, testID uuid.UUID, attachments []events.EventAttachment) error {
	_, err := e.pool.Exec(ctx, `
		DELETE FROM event_attachments WHERE event_id = $1 AND event_type = $2::event_type_enum
	`, testID, events.EventAsTest)
	if err != nil {
		return err
	}
	for _, a := range attachments {
		if a.Attachment == nil {
			continue
		}
		var attachmentID int64
		if a.AttachmentID != nil && *a.AttachmentID != 0 {
			_, err = e.pool.Exec(ctx, `
				INSERT INTO attachments (id, url, filename)
				VALUES ($1, $2, $3)
				ON CONFLICT (id) DO NOTHING
			`, *a.AttachmentID, a.Attachment.Url, a.Attachment.Filename)
			if err != nil {
				return err
			}
			attachmentID = *a.AttachmentID
		} else {
			err = e.pool.QueryRow(ctx, `
				INSERT INTO attachments (url, filename)
				VALUES ($1, $2)
				RETURNING id
			`, a.Attachment.Url, a.Attachment.Filename).Scan(&attachmentID)
			if err != nil {
				return err
			}
		}
		_, err = e.pool.Exec(ctx, `
			INSERT INTO event_attachments (event_id, event_type, attachment_id)
			VALUES ($1, $2::event_type_enum, $3)
		`, testID, events.EventAsTest, attachmentID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *eventTestRepositoryPgx) replaceEventRoles(ctx context.Context, testID uuid.UUID, roleCodes []string) error {
	_, err := e.pool.Exec(ctx, `
		DELETE FROM event_roles WHERE event_id = $1 AND event_type = 'test'::event_type_enum
	`, testID)
	if err != nil {
		return err
	}
	if len(roleCodes) == 0 {
		return nil
	}
	_, err = e.pool.Exec(ctx, `
		INSERT INTO event_roles (event_id, event_type, role_id)
		SELECT $1, 'test'::event_type_enum, r.id
		FROM roles r
		WHERE r.name = ANY($2)
		ON CONFLICT DO NOTHING
	`, testID, roleCodes)
	return err
}

func (e *eventTestRepositoryPgx) getEventRoles(ctx context.Context, testID uuid.UUID) ([]string, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT r.name
		FROM event_roles er
		JOIN roles r ON r.id = er.role_id
		WHERE er.event_id = $1 AND er.event_type = 'test'::event_type_enum
		ORDER BY r.name ASC
	`, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	codes := make([]string, 0)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, rows.Err()
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
