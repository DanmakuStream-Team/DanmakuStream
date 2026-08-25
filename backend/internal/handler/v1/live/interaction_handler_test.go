package live

import (
	"testing"

	model "danmakustream/backend/internal/model/mysql"
)

func TestSuperChatDisplaySeconds(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  int
	}{
		{name: "below first threshold", value: 49, want: 15},
		{name: "first threshold", value: 50, want: 30},
		{name: "two hundred", value: 200, want: 60},
		{name: "five hundred", value: 500, want: 90},
		{name: "one thousand", value: 1000, want: 120},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := superChatDisplaySeconds(test.value); got != test.want {
				t.Fatalf("superChatDisplaySeconds(%d) = %d, want %d", test.value, got, test.want)
			}
		})
	}
}

func TestRoomHeat(t *testing.T) {
	room := model.LiveRoom{ViewerCount: 12, LikeCount: 7, GiftValue: 230}
	const want int64 = 364
	if got := roomHeat(room); got != want {
		t.Fatalf("roomHeat() = %d, want %d", got, want)
	}
}
