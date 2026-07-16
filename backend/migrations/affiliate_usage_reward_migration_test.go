package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAffiliateUsageRewardMigrationCreatesIdempotentRewardLedger(t *testing.T) {
	sql, err := FS.ReadFile("176_affiliate_usage_rewards.sql")
	require.NoError(t, err)
	content := string(sql)
	require.Contains(t, content, "CREATE TABLE IF NOT EXISTS affiliate_usage_rewards")
	require.Contains(t, content, "UNIQUE (invitee_user_id)")
	require.Contains(t, content, "amount DECIMAL(20,8) NOT NULL")
	require.Contains(t, content, "affiliate_usage_reward_amount', '20'")
	require.Contains(t, content, "affiliate_reward_inviter_daily_limit', '5'")
	require.Contains(t, content, "affiliate_reward_inviter_30d_limit', '30'")
	require.Contains(t, content, "affiliate_reward_ip_daily_limit', '2'")
}
