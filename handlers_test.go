package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type mockTelegramServer struct {
	requests []mockRequest
	server   *httptest.Server
}

type mockRequest struct {
	method string
	path   string
	body   string
}

func newMockBot(t *testing.T) (*tgbotapi.BotAPI, *mockTelegramServer) {
	t.Helper()

	m := &mockTelegramServer{}
	m.server = httptest.NewServer(http.HandlerFunc(m.handleRequest))
	t.Cleanup(func() { m.server.Close() })

	apiEndpoint := m.server.URL + "/bot%s/%s"

	bot, err := tgbotapi.NewBotAPIWithAPIEndpoint("test-token", apiEndpoint)
	if err != nil {
		t.Fatalf("failed to create bot: %v", err)
	}
	bot.Buffer = 100

	return bot, m
}

func (m *mockTelegramServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, _ = io.ReadAll(r.Body)
		r.Body.Close()
	}

	m.requests = append(m.requests, mockRequest{
		method: r.Method,
		path:   r.URL.Path,
		body:   string(bodyBytes),
	})

	userPayload := json.RawMessage(`{"id":123,"is_bot":true,"first_name":"TestBot","username":"test_bot"}`)
	msgPayload := json.RawMessage(`{"message_id":1,"chat":{"id":456,"type":"private"},"from":{"id":789,"is_bot":true,"first_name":"TestBot"},"text":"ok"}`)

	switch {
	case strings.Contains(r.URL.Path, "/getMe"):
		writeAPIResponse(w, userPayload)
	case strings.Contains(r.URL.Path, "/sendMessage"):
		writeAPIResponse(w, msgPayload)
	case strings.Contains(r.URL.Path, "/editMessageText"):
		writeAPIResponse(w, msgPayload)
	case strings.Contains(r.URL.Path, "/answerCallbackQuery"):
		writeAPIResponse(w, json.RawMessage(`true`))
	default:
		writeAPIResponse(w, json.RawMessage(`true`))
	}
}

func writeAPIResponse(w http.ResponseWriter, result json.RawMessage) {
	json.NewEncoder(w).Encode(tgbotapi.APIResponse{
		Ok:     true,
		Result: result,
	})
}

func (m *mockTelegramServer) lastRequest() *mockRequest {
	if len(m.requests) == 0 {
		return nil
	}
	return &m.requests[len(m.requests)-1]
}

func (m *mockTelegramServer) popRequest() *mockRequest {
	if len(m.requests) == 0 {
		return nil
	}
	req := m.requests[0]
	m.requests = m.requests[1:]
	return &req
}

func (m *mockTelegramServer) reset() {
	m.requests = nil
}

func formField(body string, field string) string {
	values, err := url.ParseQuery(body)
	if err != nil {
		return ""
	}
	return values.Get(field)
}

func TestHandleStart(t *testing.T) {
	bot, mock := newMockBot(t)

	msg := &tgbotapi.Message{
		MessageID: 1,
		Chat:      &tgbotapi.Chat{ID: 456},
		Text:      "/start",
		Entities: []tgbotapi.MessageEntity{
			{Type: "bot_command", Offset: 0, Length: 6},
		},
	}

	HandleStart(bot, msg)

	req := mock.lastRequest()
	if req == nil {
		t.Fatal("expected a request, got nil")
	}
	if !strings.Contains(req.path, "sendMessage") {
		t.Errorf("expected sendMessage, got path: %s", req.path)
	}
	text := formField(req.body, "text")
	if !strings.Contains(text, "Привет") {
		t.Errorf("expected greeting text in body, got: %s", text)
	}
}

func TestHandleStatus(t *testing.T) {
	bot, mock := newMockBot(t)

	t.Run("uptime success", func(t *testing.T) {
		mock.reset()

		orig := uptimeCmd
		t.Cleanup(func() { uptimeCmd = orig })
		uptimeCmd = func() (string, error) {
			return "12:34  up 3 days,  2:15, 1 user", nil
		}

		msg := &tgbotapi.Message{
			MessageID: 1,
			Chat:      &tgbotapi.Chat{ID: 456},
			Text:      "/status",
			Entities: []tgbotapi.MessageEntity{
				{Type: "bot_command", Offset: 0, Length: 7},
			},
		}

		HandleStatus(bot, msg)

		req := mock.lastRequest()
		if req == nil {
			t.Fatal("expected a request, got nil")
		}
		text := formField(req.body, "text")
		if !strings.Contains(text, "up 3 days") {
			t.Errorf("expected uptime in text field, got: %s", text)
		}
	})

	t.Run("uptime error", func(t *testing.T) {
		mock.reset()

		orig := uptimeCmd
		t.Cleanup(func() { uptimeCmd = orig })
		uptimeCmd = func() (string, error) {
			return "", io.EOF
		}

		msg := &tgbotapi.Message{
			MessageID: 1,
			Chat:      &tgbotapi.Chat{ID: 456},
			Text:      "/status",
			Entities: []tgbotapi.MessageEntity{
				{Type: "bot_command", Offset: 0, Length: 7},
			},
		}

		HandleStatus(bot, msg)

		req := mock.lastRequest()
		if req == nil {
			t.Fatal("expected a request, got nil")
		}
		text := formField(req.body, "text")
		if !strings.Contains(text, "Ошибка") {
			t.Errorf("expected error message in text field, got: %s", text)
		}
	})
}

func TestHandleReboot(t *testing.T) {
	bot, mock := newMockBot(t)

	msg := &tgbotapi.Message{
		MessageID: 1,
		Chat:      &tgbotapi.Chat{ID: 456},
		Text:      "/reboot",
		Entities: []tgbotapi.MessageEntity{
			{Type: "bot_command", Offset: 0, Length: 7},
		},
	}

	HandleReboot(bot, msg)

	req := mock.lastRequest()
	if req == nil {
		t.Fatal("expected a request, got nil")
	}
	text := formField(req.body, "text")
	if !strings.Contains(text, "перезагрузить") {
		t.Errorf("expected confirmation text in body, got: %s", text)
	}
	markup := formField(req.body, "reply_markup")
	if !strings.Contains(markup, callbackRebootConfirm) {
		t.Errorf("expected %s in reply_markup, got: %s", callbackRebootConfirm, markup)
	}
	if !strings.Contains(markup, callbackRebootCancel) {
		t.Errorf("expected %s in reply_markup, got: %s", callbackRebootCancel, markup)
	}
}

func TestHandleCallback(t *testing.T) {
	bot, mock := newMockBot(t)

	t.Run("confirm", func(t *testing.T) {
		mock.reset()

		origReboot := rebootCmd
		t.Cleanup(func() { rebootCmd = origReboot })
		rebootCmd = func() error { return nil }

		cb := &tgbotapi.CallbackQuery{
			ID:   "cb-1",
			Data: callbackRebootConfirm,
			From: &tgbotapi.User{ID: 123, UserName: "testuser"},
			Message: &tgbotapi.Message{
				MessageID: 5,
				Chat:      &tgbotapi.Chat{ID: 456},
				Text:      "Вы уверены?",
			},
		}

		HandleCallback(bot, cb)

		if len(mock.requests) < 2 {
			t.Fatalf("expected at least 2 requests (callback answer + edit), got %d", len(mock.requests))
		}

		callbackReq := mock.popRequest()
		if !strings.Contains(callbackReq.path, "answerCallbackQuery") {
			t.Errorf("expected answerCallbackQuery first, got: %s", callbackReq.path)
		}

		editReq := mock.popRequest()
		if !strings.Contains(editReq.path, "editMessageText") {
			t.Errorf("expected editMessageText, got: %s", editReq.path)
		}
		text := formField(editReq.body, "text")
		if !strings.Contains(text, "Перезагружаю") {
			t.Errorf("expected rebooting text, got: %s", text)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		mock.reset()

		cb := &tgbotapi.CallbackQuery{
			ID:   "cb-2",
			Data: callbackRebootCancel,
			From: &tgbotapi.User{ID: 123, UserName: "testuser"},
			Message: &tgbotapi.Message{
				MessageID: 5,
				Chat:      &tgbotapi.Chat{ID: 456},
				Text:      "Вы уверены?",
			},
		}

		HandleCallback(bot, cb)

		if len(mock.requests) < 2 {
			t.Fatalf("expected at least 2 requests, got %d", len(mock.requests))
		}

		callbackReq := mock.popRequest()
		if !strings.Contains(callbackReq.path, "answerCallbackQuery") {
			t.Errorf("expected answerCallbackQuery first, got: %s", callbackReq.path)
		}

		editReq := mock.popRequest()
		if !strings.Contains(editReq.path, "editMessageText") {
			t.Errorf("expected editMessageText, got: %s", editReq.path)
		}
		text := formField(editReq.body, "text")
		if !strings.Contains(text, "отменена") {
			t.Errorf("expected cancelled text, got: %s", text)
		}
	})
}
