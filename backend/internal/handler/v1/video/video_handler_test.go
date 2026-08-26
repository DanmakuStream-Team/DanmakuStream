package video

import "testing"

// UNIT-TC13-06 审核状态校验：pending/approved/rejected 合法，其余非法
func TestIsValidVideoStatus(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{status: "pending", want: true},
		{status: "approved", want: true},
		{status: "rejected", want: true},
		{status: "", want: false},
		{status: "publish", want: false},
		{status: "Approved", want: false},
		{status: "approved ", want: false},
		{status: "deleted", want: false},
	}
	for _, tc := range cases {
		if got := isValidVideoStatus(tc.status); got != tc.want {
			t.Errorf("isValidVideoStatus(%q) = %v, want %v", tc.status, got, tc.want)
		}
	}
}
