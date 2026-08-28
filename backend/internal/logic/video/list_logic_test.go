package videologic

import "testing"

// UNIT-TC02-01 视频发现排序：仅支持产品约定的四种排序方式。
func TestVideoSortExpr(t *testing.T) {
	const hot = "hot_score"
	cases := []struct {
		name    string
		sort    string
		want    string
		wantErr bool
	}{
		{name: "empty_defaults_to_hot", sort: "", want: hot + " DESC, created_at DESC"},
		{name: "hot", sort: "hot", want: hot + " DESC, created_at DESC"},
		{name: "date", sort: "date", want: "created_at DESC"},
		{name: "like", sort: "like", want: "like_count DESC, created_at DESC"},
		{name: "collect", sort: "collect", want: "collect_count DESC, created_at DESC"},
		{name: "invalid", sort: "views", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := videoSortExpr(tc.sort, hot)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("videoSortExpr(%q) expected error, got %q", tc.sort, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("videoSortExpr(%q) unexpected error: %v", tc.sort, err)
			}
			if got != tc.want {
				t.Errorf("videoSortExpr(%q) = %q, want %q", tc.sort, got, tc.want)
			}
		})
	}
}
