package chat

import (
	"errors"
	"testing"
)

func TestValidateMessageInput(t *testing.T) {
	tests := []struct {
		name      string
		input     CreateMessageInput
		wantError error
		wantText  string
	}{
		{name: "text", input: CreateMessageInput{Type: MessageTypeText, Content: " hello "}},
		{name: "empty text", input: CreateMessageInput{Type: MessageTypeText}, wantError: ErrEmptyContent},
		{name: "owned image", input: CreateMessageInput{Type: MessageTypeImage, MediaURL: "/media/messages/7/20260801/a.png"}, wantText: "[图片]"},
		{name: "other user image", input: CreateMessageInput{Type: MessageTypeImage, MediaURL: "/media/messages/8/20260801/a.png"}, wantError: ErrInvalidMedia},
		{name: "wrong media extension", input: CreateMessageInput{Type: MessageTypeVideo, MediaURL: "/media/messages/7/20260801/a.jpg"}, wantError: ErrInvalidMedia},
		{name: "video share", input: CreateMessageInput{Type: MessageTypeVideoShare, VideoID: 3}, wantText: "[视频分享]"},
		{name: "missing shared video", input: CreateMessageInput{Type: MessageTypeVideoShare}, wantError: ErrVideoMissing},
		{name: "unknown type", input: CreateMessageInput{Type: "file"}, wantError: ErrInvalidType},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := test.input
			err := validateMessageInput(7, &input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("validateMessageInput() error = %v, want %v", err, test.wantError)
			}
			if test.wantText != "" && input.Content != test.wantText {
				t.Fatalf("content = %q, want %q", input.Content, test.wantText)
			}
		})
	}
}

func TestNormalizedMessageType(t *testing.T) {
	if got := normalizedMessageType(""); got != MessageTypeText {
		t.Fatalf("normalizedMessageType(\"\") = %q", got)
	}
	if got := normalizedMessageType(MessageTypeImage); got != MessageTypeImage {
		t.Fatalf("normalizedMessageType(image) = %q", got)
	}
}
