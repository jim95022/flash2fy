package telegram

const (
	messageUsage       = "Send any text message to create a card with that text on the front. Back will be empty. Use /help for this hint."
	messageUnknownCmd  = "Unknown command. " + messageUsage
	messageEmptyIgnore = "Empty cards are ignored. " + messageUsage
	messageCreateOK    = "Card created ✅\nID: %s\nFront: %s\nBack: %s"
	messageCreateFail  = "Failed to create card: %v"
)
