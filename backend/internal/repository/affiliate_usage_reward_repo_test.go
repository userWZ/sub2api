package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestEvaluateAffiliateUsageRewardRisk_MissingIPRequiresReview(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	reason, err := evaluateAffiliateUsageRewardRisk(context.Background(), db, 1, 2, "", service.AffiliateSettings{
		RewardReviewMissingIP: true,
	})
	require.NoError(t, err)
	require.Equal(t, "missing_ip", reason)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEvaluateAffiliateUsageRewardRisk_SameIPRequiresReview(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`(?s)SELECT ip_address.*FROM usage_logs.*WHERE user_id = \$1`).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"ip_address"}).AddRow("203.0.113.10"))
	reason, err := evaluateAffiliateUsageRewardRisk(context.Background(), db, 1, 2, "203.0.113.10", service.AffiliateSettings{
		RewardReviewSameIP: true,
	})
	require.NoError(t, err)
	require.Equal(t, "same_ip", reason)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEvaluateAffiliateUsageRewardRisk_InviterDailyLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`(?s)SELECT COUNT\(\*\).*FROM affiliate_usage_rewards.*inviter_user_id = \$1.*24 hours`).
		WithArgs(int64(2), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	reason, err := evaluateAffiliateUsageRewardRisk(context.Background(), db, 1, 2, "203.0.113.20", service.AffiliateSettings{
		RewardInviterDailyLimit: 5,
	})
	require.NoError(t, err)
	require.Equal(t, "inviter_daily_limit", reason)
	require.NoError(t, mock.ExpectationsWereMet())
}
