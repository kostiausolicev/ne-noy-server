package service

import (
	"context"
	"ne_noy/internal/client"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type notificationService struct {
	pgx      *pgxpool.Pool
	vkClient client.VkApiClient
}

type userRef struct {
	id   uuid.UUID
	vkID int64
}

func (n *notificationService) SendNotificationForUser(ctx context.Context, destination uuid.UUID, chanel, message, fragment string) error {
	return n.SendNotificationForUsers(ctx, []uuid.UUID{destination}, chanel, message, fragment)
}

func (n *notificationService) SendNotificationForUsers(ctx context.Context, destination []uuid.UUID, chanel, message, fragment string) error {
	users, err := n.fetchUserRefs(ctx, `SELECT id, vk_id FROM users WHERE users.notification_available = true AND id = ANY($1)`, destination)
	if err != nil {
		return err
	}

	var notificationID uuid.UUID
	err = n.pgx.QueryRow(ctx, `
		INSERT INTO notifications (title, fragment, forall)
		VALUES ($1, $2, $3)
		RETURNING id`, message, fragment, false).Scan(&notificationID)
	if err != nil {
		return err
	}

	return n.sendParallelMessages(ctx, users, message, fragment, notificationID)
}

func (n *notificationService) SendNotificationForAll(ctx context.Context, chanel, message, fragment string) error {
	users, err := n.fetchUserRefs(ctx, `SELECT id, vk_id FROM users WHERE users.notification_available = true`)
	if err != nil {
		return err
	}

	var notificationID uuid.UUID
	err = n.pgx.QueryRow(ctx, `
		INSERT INTO notifications (title, fragment, forall)
		VALUES ($1, $2, $3)
		RETURNING id`, message, fragment, true).Scan(&notificationID)
	if err != nil {
		return err
	}

	return n.sendParallelMessages(ctx, users, message, fragment, notificationID)
}

func (n *notificationService) SendNotificationForRole(ctx context.Context, role, chanel, message, fragment string) error {
	var roleID uuid.UUID
	if err := n.pgx.QueryRow(ctx, `SELECT id FROM roles WHERE name = $1`, role).Scan(&roleID); err != nil {
		return err
	}

	users, err := n.fetchUserRefs(ctx, `SELECT id, vk_id FROM users WHERE users.notification_available = true AND  role_id = $1`, roleID)
	if err != nil {
		return err
	}

	var notificationID uuid.UUID
	err = n.pgx.QueryRow(ctx, `
		INSERT INTO notifications (title, fragment, forall, userrole)
		VALUES ($1, $2, $3, $4)
		RETURNING id`, message, fragment, true, roleID).Scan(&notificationID)
	if err != nil {
		return err
	}

	return n.sendParallelMessages(ctx, users, message, fragment, notificationID)
}

// fetchUserRefs выполняет запрос с произвольными аргументами и возвращает пары (uuid, vk_id).
func (n *notificationService) fetchUserRefs(ctx context.Context, query string, args ...any) ([]userRef, error) {
	rows, err := n.pgx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []userRef
	for rows.Next() {
		var u userRef
		if err := rows.Scan(&u.id, &u.vkID); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (n *notificationService) sendParallelMessages(ctx context.Context, users []userRef, message, fragment string, notificationID uuid.UUID) error {
	if len(users) == 0 {
		return nil
	}

	const batchSize = 10

	type job struct {
		users []userRef
	}

	var wg sync.WaitGroup
	workersCount := 10
	jobs := make(chan job, workersCount)
	errs := make(chan error, 1)

	taskCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for w := 0; w < workersCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				vkIDs := make([]int64, len(j.users))
				for i, u := range j.users {
					vkIDs[i] = u.vkID
				}

				if _, err := n.vkClient.SendNotificationForUsers(vkIDs, message, fragment); err != nil {
					select {
					case errs <- err:
						cancel()
					default:
					}
					return
				}

				for _, u := range j.users {
					if _, err := n.pgx.Exec(taskCtx, `
						INSERT INTO notification_user (notificationid, userid) VALUES ($1, $2)
					`, notificationID, u.id); err != nil {
						select {
						case errs <- err:
							cancel()
						default:
						}
						return
					}
				}
			}
		}()
	}

sendJobs:
	for i := 0; i < len(users); i += batchSize {
		right := i + batchSize
		if right > len(users) {
			right = len(users)
		}
		select {
		case <-taskCtx.Done():
			break sendJobs
		case jobs <- job{users: users[i:right]}:
		}
	}
	close(jobs)
	wg.Wait()

	select {
	case err := <-errs:
		return err
	default:
		return nil
	}
}

type NotificationService interface {
	SendNotificationForUser(ctx context.Context, destination uuid.UUID, chanel, message, fragment string) error
	SendNotificationForUsers(ctx context.Context, destination []uuid.UUID, chanel, message, fragment string) error
	SendNotificationForAll(ctx context.Context, chanel, message, fragment string) error
	SendNotificationForRole(ctx context.Context, role, chanel, message, fragment string) error
}

func NewNotificationService(pgx *pgxpool.Pool, vkClient client.VkApiClient) NotificationService {
	return &notificationService{
		pgx:      pgx,
		vkClient: vkClient,
	}
}
