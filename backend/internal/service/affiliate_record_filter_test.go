package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeAffiliateRecordFilterAction(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "discount", in: "discount", want: "discount"},
		{name: "discount restore", in: "discount_restore", want: "discount_restore"},
		{name: "withdraw", in: "withdraw", want: "withdraw"},
		{name: "trim spaces", in: " withdraw ", want: "withdraw"},
		{name: "legacy transfer hidden", in: "transfer", want: ""},
		{name: "unknown action", in: "accrue", want: ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeAffiliateRecordFilter(AffiliateRecordFilter{Action: tc.in})
			require.Equal(t, tc.want, got.Action)
		})
	}
}
