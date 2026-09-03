package chat

import (
	"context"
	"errors"
	"strings"
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
		{name: "text too long", input: CreateMessageInput{Type: MessageTypeText, Content: strings.Repeat("x", 2001)}, wantError: ErrTooLong},
		{name: "media name too long", input: CreateMessageInput{Type: MessageTypeImage, MediaURL: "/media/messages/7/20260801/a.png", MediaName: strings.Repeat("x", 201)}, wantError: ErrTooLong},
		{name: "path traversal", input: CreateMessageInput{Type: MessageTypeImage, MediaURL: "/media/messages/7/../a.png"}, wantError: ErrInvalidMedia},
		{name: "query string", input: CreateMessageInput{Type: MessageTypeImage, MediaURL: "/media/messages/7/a.png?token=x"}, wantError: ErrInvalidMedia},
		{name: "owned video", input: CreateMessageInput{Type: MessageTypeVideo, MediaURL: "/media/messages/7/20260801/a.mp4"}, wantText: "[视频]"},
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

func TestCreateAndBroadcastRejectsSelfBeforeDatabase(t *testing.T) {
	hub := &Hub{}
	_, err := hub.CreateAndBroadcast(context.Background(), "", 7, CreateMessageInput{ReceiverID: 7, Type: MessageTypeText, Content: "hello"})
	if !errors.Is(err, ErrSelfMessage) {
		t.Fatalf("error = %v, want %v", err, ErrSelfMessage)
	}
}

func TestCreateAndBroadcastRejectsLongClientMessageIDBeforeDatabase(t *testing.T) {
	hub := &Hub{}
	_, err := hub.CreateAndBroadcast(context.Background(), "", 7, CreateMessageInput{
		ReceiverID: 8, Type: MessageTypeText, Content: "hello", ClientMessageID: strings.Repeat("x", 65),
	})
	if !errors.Is(err, ErrInvalidClientMessageID) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidClientMessageID)
	}
}

func TestChatErrorMessage(t *testing.T) {
	for _, test := range []struct {
		err  error
		want string
	}{
		{ErrEmptyContent, "消息内容不能为空"},
		{ErrTooLong, "消息不能超过 2000 个字符"},
		{ErrSelfMessage, "不能给自己发送私信"},
		{ErrUserNotFound, "接收用户不存在"},
		{ErrBlocked, "存在拉黑关系，无法发送私信"},
		{ErrInvalidType, "不支持的消息类型"},
		{ErrInvalidMedia, "私信附件无效，请重新上传"},
		{ErrVideoMissing, "分享的视频不存在或尚未公开"},
		{ErrInvalidClientMessageID, "客户端消息编号无效"},
		{errors.New("unknown"), "消息发送失败"},
	} {
		if got := ErrorMessage(test.err); got != test.want {
			t.Fatalf("error %v => %q, want %q", test.err, got, test.want)
		}
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
