package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *affiliateRepository) ProcessUsageReward(ctx context.Context, event service.AffiliateUsageRewardEvent, settings service.AffiliateSettings) (*service.AffiliateUsageRewardProcessResult, error) {
	if event.InviteeUserID <= 0 || event.ActualCost <= 0 || settings.UsageRewardAmount <= 0 {
		return nil, nil
	}
	result := &service.AffiliateUsageRewardProcessResult{}
	err := r.withTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		var inviterID int64
		err := scanAffiliateUsageRewardRow(txCtx, txClient, `
SELECT inviter_id
FROM user_affiliates
WHERE user_id = $1 AND inviter_id IS NOT NULL
LIMIT 1`, []any{event.InviteeUserID}, &inviterID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("find inviter for usage reward: %w", err)
		}
		if inviterID <= 0 || inviterID == event.InviteeUserID {
			return nil
		}

		triggerIP := strings.TrimSpace(event.IPAddress)
		// Serialize rewards per inviter so concurrent first-use events cannot both
		// pass a count threshold. The advisory lock adds the same protection for
		// one IP shared across different inviters.
		var lockedInviterID int64
		if err := scanAffiliateUsageRewardRow(txCtx, txClient, `
SELECT user_id FROM user_affiliates WHERE user_id = $1 FOR UPDATE`, []any{inviterID}, &lockedInviterID); err != nil {
			return fmt.Errorf("lock affiliate inviter for usage reward: %w", err)
		}
		if triggerIP != "" {
			if _, err := txClient.ExecContext(txCtx, `SELECT pg_advisory_xact_lock(hashtext($1))`, triggerIP); err != nil {
				return fmt.Errorf("lock affiliate reward trigger ip: %w", err)
			}
		}
		var rewardID int64
		err = scanAffiliateUsageRewardRow(txCtx, txClient, `
INSERT INTO affiliate_usage_rewards (
    inviter_user_id, invitee_user_id, amount, status, risk_reason,
    trigger_request_id, trigger_api_key_id, trigger_ip, trigger_actual_cost,
    created_at, updated_at
)
VALUES ($1, $2, $3, 'pending', 'evaluating', $4, NULLIF($5, 0), NULLIF($6, ''), $7, NOW(), NOW())
ON CONFLICT (invitee_user_id) DO NOTHING
RETURNING id`, []any{inviterID, event.InviteeUserID, settings.UsageRewardAmount,
			event.RequestID, event.APIKeyID, triggerIP, event.ActualCost}, &rewardID)
		if errors.Is(err, sql.ErrNoRows) {
			result.Created = false
			return nil
		}
		if err != nil {
			return fmt.Errorf("create affiliate usage reward: %w", err)
		}

		riskReason, err := evaluateAffiliateUsageRewardRisk(txCtx, txClient, rewardID, inviterID, triggerIP, settings)
		if err != nil {
			return err
		}
		status := "pending"
		if riskReason == "" {
			status = "granted"
			riskReason = "auto_granted"
			res, err := txClient.ExecContext(txCtx, `
UPDATE users
SET balance = balance + $1, updated_at = NOW()
WHERE id = $2 AND deleted_at IS NULL`, settings.UsageRewardAmount, inviterID)
			if err != nil {
				return fmt.Errorf("grant affiliate usage reward balance: %w", err)
			}
			affected, err := res.RowsAffected()
			if err != nil {
				return err
			}
			if affected == 0 {
				return service.ErrUserNotFound
			}
		}

		if _, err := txClient.ExecContext(txCtx, `
UPDATE affiliate_usage_rewards
SET status = $1, risk_reason = $2, updated_at = NOW()
WHERE id = $3`, status, riskReason, rewardID); err != nil {
			return fmt.Errorf("finalize affiliate usage reward: %w", err)
		}
		result = &service.AffiliateUsageRewardProcessResult{
			Created:    true,
			RewardID:   rewardID,
			InviterID:  inviterID,
			Status:     status,
			RiskReason: riskReason,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func evaluateAffiliateUsageRewardRisk(
	ctx context.Context,
	client affiliateQueryExecer,
	rewardID int64,
	inviterID int64,
	triggerIP string,
	settings service.AffiliateSettings,
) (string, error) {
	if triggerIP == "" && settings.RewardReviewMissingIP {
		return "missing_ip", nil
	}
	if triggerIP != "" && settings.RewardReviewSameIP {
		var inviterIP string
		rows, err := client.QueryContext(ctx, `
SELECT ip_address
FROM usage_logs
WHERE user_id = $1
  AND actual_cost > 0
  AND ip_address IS NOT NULL
  AND ip_address <> ''
ORDER BY created_at DESC
LIMIT 1`, inviterID)
		if err != nil {
			return "", fmt.Errorf("query inviter latest ip: %w", err)
		}
		if rows.Next() {
			if scanErr := rows.Scan(&inviterIP); scanErr != nil {
				_ = rows.Close()
				return "", scanErr
			}
		}
		if closeErr := rows.Close(); closeErr != nil {
			return "", closeErr
		}
		if strings.TrimSpace(inviterIP) != "" && strings.TrimSpace(inviterIP) == triggerIP {
			return "same_ip", nil
		}
	}

	if settings.RewardInviterDailyLimit > 0 {
		count, err := scanInt64(ctx, client, `
SELECT COUNT(*)
FROM affiliate_usage_rewards
WHERE inviter_user_id = $1 AND id <> $2 AND created_at >= NOW() - INTERVAL '24 hours'`, inviterID, rewardID)
		if err != nil {
			return "", err
		}
		if count >= int64(settings.RewardInviterDailyLimit) {
			return "inviter_daily_limit", nil
		}
	}
	if settings.RewardInviterRolling30DayLimit > 0 {
		count, err := scanInt64(ctx, client, `
SELECT COUNT(*)
FROM affiliate_usage_rewards
WHERE inviter_user_id = $1 AND id <> $2 AND created_at >= NOW() - INTERVAL '30 days'`, inviterID, rewardID)
		if err != nil {
			return "", err
		}
		if count >= int64(settings.RewardInviterRolling30DayLimit) {
			return "inviter_30d_limit", nil
		}
	}
	if triggerIP != "" && settings.RewardIPDailyLimit > 0 {
		count, err := scanInt64(ctx, client, `
SELECT COUNT(*)
FROM affiliate_usage_rewards
WHERE trigger_ip = $1 AND id <> $2 AND created_at >= NOW() - INTERVAL '24 hours'`, triggerIP, rewardID)
		if err != nil {
			return "", err
		}
		if count >= int64(settings.RewardIPDailyLimit) {
			return "ip_daily_limit", nil
		}
	}
	return "", nil
}

func (r *affiliateRepository) ListUsageRewardRecords(ctx context.Context, filter service.AffiliateUsageRewardFilter) ([]service.AffiliateUsageRewardRecord, int64, error) {
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	like := "%" + strings.TrimSpace(filter.Search) + "%"
	args := []any{like}
	where := []string{"(inviter.email ILIKE $1 OR inviter.username ILIKE $1 OR invitee.email ILIKE $1 OR invitee.username ILIKE $1)"}
	if filter.Status != "" {
		args = append(args, filter.Status)
		where = append(where, fmt.Sprintf("reward.status = $%d", len(args)))
	}
	if filter.StartAt != nil {
		args = append(args, *filter.StartAt)
		where = append(where, fmt.Sprintf("reward.created_at >= $%d", len(args)))
	}
	if filter.EndAt != nil {
		args = append(args, *filter.EndAt)
		where = append(where, fmt.Sprintf("reward.created_at <= $%d", len(args)))
	}
	from := `
FROM affiliate_usage_rewards reward
JOIN users inviter ON inviter.id = reward.inviter_user_id
JOIN users invitee ON invitee.id = reward.invitee_user_id
LEFT JOIN users reviewer ON reviewer.id = reward.reviewer_user_id
WHERE ` + strings.Join(where, " AND ")
	client := clientFromContext(ctx, r.client)
	total, err := scanInt64(ctx, client, "SELECT COUNT(*) "+from, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("count affiliate usage rewards: %w", err)
	}
	args = append(args, pageSize, offset)
	query := `
SELECT reward.id,
       reward.inviter_user_id, COALESCE(inviter.email, ''), COALESCE(inviter.username, ''),
       reward.invitee_user_id, COALESCE(invitee.email, ''), COALESCE(invitee.username, ''),
       reward.amount::double precision, reward.status, reward.risk_reason,
       reward.trigger_request_id, reward.trigger_ip, reward.trigger_actual_cost::double precision,
       reward.reviewer_user_id, reviewer.email, reward.review_remark, reward.reviewed_at, reward.created_at
` + from + fmt.Sprintf(" ORDER BY reward.created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := client.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list affiliate usage rewards: %w", err)
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.AffiliateUsageRewardRecord, 0, pageSize)
	for rows.Next() {
		var item service.AffiliateUsageRewardRecord
		var triggerIP sql.NullString
		var reviewerID sql.NullInt64
		var reviewerEmail sql.NullString
		var reviewedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.InviterID, &item.InviterEmail, &item.InviterUsername,
			&item.InviteeID, &item.InviteeEmail, &item.InviteeUsername,
			&item.Amount, &item.Status, &item.RiskReason,
			&item.TriggerRequestID, &triggerIP, &item.TriggerActualCost,
			&reviewerID, &reviewerEmail, &item.ReviewRemark, &reviewedAt, &item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		if triggerIP.Valid {
			item.TriggerIPAddress = &triggerIP.String
		}
		if reviewerID.Valid {
			item.ReviewerUserID = &reviewerID.Int64
		}
		if reviewerEmail.Valid {
			item.ReviewerEmail = &reviewerEmail.String
		}
		if reviewedAt.Valid {
			item.ReviewedAt = &reviewedAt.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *affiliateRepository) ReviewUsageReward(ctx context.Context, rewardID, reviewerUserID int64, approve bool, remark string) (*service.AffiliateUsageRewardReviewResult, error) {
	result := &service.AffiliateUsageRewardReviewResult{RewardID: rewardID}
	err := r.withTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		var status string
		if err := scanAffiliateUsageRewardRow(txCtx, txClient, `
SELECT inviter_user_id, amount::double precision, status
FROM affiliate_usage_rewards
WHERE id = $1
FOR UPDATE`, []any{rewardID}, &result.InviterID, &result.Amount, &status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return service.ErrAffiliateUsageRewardNotFound
			}
			return err
		}
		if status != "pending" {
			return service.ErrAffiliateUsageRewardReviewed
		}
		result.Status = "rejected"
		reason := "manual_rejected"
		if approve {
			result.Status = "granted"
			reason = "manual_approved"
			res, err := txClient.ExecContext(txCtx, `
UPDATE users
SET balance = balance + $1, updated_at = NOW()
WHERE id = $2 AND deleted_at IS NULL`, result.Amount, result.InviterID)
			if err != nil {
				return err
			}
			affected, err := res.RowsAffected()
			if err != nil {
				return err
			}
			if affected == 0 {
				return service.ErrUserNotFound
			}
		}
		_, err := txClient.ExecContext(txCtx, `
UPDATE affiliate_usage_rewards
SET status = $1,
    risk_reason = $2,
    reviewer_user_id = $3,
    review_remark = $4,
    reviewed_at = NOW(),
    updated_at = NOW()
WHERE id = $5`, result.Status, reason, reviewerUserID, strings.TrimSpace(remark), rewardID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func scanAffiliateUsageRewardRow(ctx context.Context, client affiliateQueryExecer, query string, args []any, dest ...any) error {
	rows, err := client.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	if err := rows.Scan(dest...); err != nil {
		return err
	}
	return rows.Err()
}
