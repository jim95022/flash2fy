package cardbot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *updateHandler) dispatch(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	text := strings.TrimSpace(update.Message.Text)

	if strings.HasPrefix(text, "/") {
		h.handleCommand(ctx, b, chatID, text)
		return
	}

	h.handleCreateCard(ctx, b, update, text)
}

func (h *updateHandler) handleCommand(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	command, _ := splitCommand(text)
	switch command {
	case "/start", "/help":
		h.sendMessage(ctx, b, chatID, messageUsage)
	default:
		h.sendMessage(ctx, b, chatID, messageUnknownCmd)
	}
}

func (h *updateHandler) handleCreateCard(ctx context.Context, b *bot.Bot, update *models.Update, message string) {
	chatID := update.Message.Chat.ID
	front := strings.TrimSpace(message)
	if front == "" {
		h.sendMessage(ctx, b, chatID, messageEmptyIgnore)
		return
	}

	ownerID := resolveOwnerID(update)
	card, err := h.createCard(front, "", ownerID)
	if err != nil {
		h.sendMessage(ctx, b, chatID, fmt.Sprintf(messageCreateFail, err))
		return
	}

	response := fmt.Sprintf(messageCreateOK, card.ID, card.Front, card.Back)
	h.sendMessage(ctx, b, chatID, response)
}

func (h *updateHandler) sendMessage(ctx context.Context, b *bot.Bot, chatID int64, message string) {
	if err := h.send(ctx, b, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   message,
	}); err != nil {
		log.Printf("telegram: failed sending message: %v", err)
	}
}

func splitCommand(text string) (cmd string, payload string) {
	if text == "" {
		return "", ""
	}
	trimmed := strings.TrimSpace(text)
	sep := strings.IndexAny(trimmed, " \n")
	if sep == -1 {
		return trimmed, ""
	}
	cmd = trimmed[:sep]
	payload = strings.TrimSpace(trimmed[sep+1:])
	if i := strings.Index(cmd, "@"); i != -1 {
		cmd = cmd[:i]
	}
	return cmd, payload
}

func resolveOwnerID(update *models.Update) string {
	if update.Message != nil && update.Message.From != nil {
		return strconv.FormatInt(update.Message.From.ID, 10)
	}
	if update.Message != nil {
		return strconv.FormatInt(update.Message.Chat.ID, 10)
	}
	return ""
}
