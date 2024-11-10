package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	db "bitbucket.org/jessyw/go_simplebank/db/sqlc"
	"bitbucket.org/jessyw/go_simplebank/util"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskSendVerifyEmail = "task:send_verify_email"

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)

	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("queue", info.Queue).
		Int("max_retry", info.MaxRetry).
		Int("retryied", info.Retried).
		Msg("enqueued task")

	return err

}

func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("user doesn't exist: %w", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get user: %w", asynq.SkipRetry)
	}

	verifyEmail, err := processor.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: util.RandomString(32),
	})

	if err != nil {
		return fmt.Errorf("failed to create verify email: %w", asynq.SkipRetry)
	}

	config, err := util.LoadConfig("..")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", asynq.SkipRetry)
	}

	verifyUrl := fmt.Sprintf("%s?email_id=%d&secret_code=%s", config.VerifyEmailUrl,
		verifyEmail.ID, verifyEmail.SecretCode)

	subject := "Welcome to Simple Bank"
	body := fmt.Sprintf(`Hello %s, <br/>
	Thank you for registering with us!<br/>
	Please <a href="%s">click here</a> to verify your email address.<br/>
	`, user.FullName, verifyUrl)

	to := []string{user.Email}

	err = processor.mailer.SendEmail(
		to,
		nil,
		nil,
		subject,
		body,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("username", user.Username).
		Str("email", user.Email).
		Msg("processed task send verify email")

	return nil
}
