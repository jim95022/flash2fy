package cardbot

import "testing"

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantCmd  string
		wantBody string
	}{
		{"simple", "/add front | back", "/add", "front | back"},
		{"newline", "/add\nfront\nback", "/add", "front\nback"},
		{"with mention", "/add@flash2fy hello | world", "/add", "hello | world"},
		{"no payload", "/start", "/start", ""},
		{"leading spaces", "   /help   please", "/help", "please"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, body := splitCommand(tc.input)
			if cmd != tc.wantCmd || body != tc.wantBody {
				t.Fatalf("splitCommand(%q) = %q, %q; want %q, %q", tc.input, cmd, body, tc.wantCmd, tc.wantBody)
			}
		})
	}
}
