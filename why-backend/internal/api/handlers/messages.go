package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"why-backend/internal/models"
)

var messageTracer = otel.Tracer("why-backend/handlers/messages")

type MessageHandler struct {
	db *sql.DB
}

func NewMessageHandler(db *sql.DB) *MessageHandler {
	return &MessageHandler{db: db}
}

// CreateMessage creates a new message
func (h *MessageHandler) CreateMessage(c *gin.Context) {
	ctx, span := messageTracer.Start(c.Request.Context(), "CreateMessage")
	defer span.End()

	userID, _ := c.Get("user_id")
	span.SetAttributes(attribute.String("user.id", userID.(string)))

	var req models.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var message models.Message
	err := h.db.QueryRowContext(ctx,
		`INSERT INTO messages (user_id, content, media_urls)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, content, media_urls, created_at, updated_at`,
		userID, req.Content, pq.Array(req.MediaURLs),
	).Scan(&message.ID, &message.UserID, &message.Content, &message.MediaURLs, &message.CreatedAt, &message.UpdatedAt)

	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to create message", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create message"})
		return
	}

	span.SetAttributes(
		attribute.String("message.id", message.ID),
		attribute.Int("media_urls.count", len(req.MediaURLs)),
	)
	slog.InfoContext(ctx, "Message created", "message_id", message.ID, "user_id", userID)

	c.JSON(http.StatusCreated, message)
}

// ListMessages returns paginated messages
func (h *MessageHandler) ListMessages(c *gin.Context) {
	ctx, span := messageTracer.Start(c.Request.Context(), "ListMessages")
	defer span.End()

	rows, err := h.db.QueryContext(ctx,
		`SELECT id, user_id, content, media_urls, created_at, updated_at
		 FROM messages
		 ORDER BY created_at DESC
		 LIMIT 50`,
	)
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to list messages", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list messages"})
		return
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.Content, &msg.MediaURLs, &msg.CreatedAt, &msg.UpdatedAt); err != nil {
			span.RecordError(err)
			slog.ErrorContext(ctx, "Failed to scan message", "error", err)
			continue
		}
		messages = append(messages, msg)
	}

	span.SetAttributes(attribute.Int("messages.count", len(messages)))
	c.JSON(http.StatusOK, messages)
}

// GetMessage returns a single message with its replies
func (h *MessageHandler) GetMessage(c *gin.Context) {
	ctx, span := messageTracer.Start(c.Request.Context(), "GetMessage")
	defer span.End()

	messageID := c.Param("id")
	span.SetAttributes(attribute.String("message.id", messageID))

	var message models.Message
	err := h.db.QueryRowContext(ctx,
		`SELECT id, user_id, content, media_urls, created_at, updated_at
		 FROM messages WHERE id = $1`,
		messageID,
	).Scan(&message.ID, &message.UserID, &message.Content, &message.MediaURLs, &message.CreatedAt, &message.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	} else if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to get message", "error", err, "message_id", messageID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get message"})
		return
	}

	c.JSON(http.StatusOK, message)
}

// CreateReply creates a reply to a message
func (h *MessageHandler) CreateReply(c *gin.Context) {
	ctx, span := messageTracer.Start(c.Request.Context(), "CreateReply")
	defer span.End()

	messageID := c.Param("id")
	userID, _ := c.Get("user_id")

	span.SetAttributes(
		attribute.String("message.id", messageID),
		attribute.String("user.id", userID.(string)),
	)

	var req models.CreateReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var reply models.Reply
	err := h.db.QueryRowContext(ctx,
		`INSERT INTO replies (message_id, user_id, content, media_urls)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, message_id, user_id, content, media_urls, created_at, updated_at`,
		messageID, userID, req.Content, pq.Array(req.MediaURLs),
	).Scan(&reply.ID, &reply.MessageID, &reply.UserID, &reply.Content, &reply.MediaURLs, &reply.CreatedAt, &reply.UpdatedAt)

	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to create reply", "error", err, "message_id", messageID, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create reply"})
		return
	}

	span.SetAttributes(attribute.String("reply.id", reply.ID))
	slog.InfoContext(ctx, "Reply created", "reply_id", reply.ID, "message_id", messageID, "user_id", userID)

	c.JSON(http.StatusCreated, reply)
}

// ListReplies returns all replies for a message
func (h *MessageHandler) ListReplies(c *gin.Context) {
	ctx, span := messageTracer.Start(c.Request.Context(), "ListReplies")
	defer span.End()

	messageID := c.Param("id")
	span.SetAttributes(attribute.String("message.id", messageID))

	rows, err := h.db.QueryContext(ctx,
		`SELECT id, message_id, user_id, content, media_urls, created_at, updated_at
		 FROM replies
		 WHERE message_id = $1
		 ORDER BY created_at ASC`,
		messageID,
	)
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to list replies", "error", err, "message_id", messageID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list replies"})
		return
	}
	defer rows.Close()

	var replies []models.Reply
	for rows.Next() {
		var reply models.Reply
		if err := rows.Scan(&reply.ID, &reply.MessageID, &reply.UserID, &reply.Content, &reply.MediaURLs, &reply.CreatedAt, &reply.UpdatedAt); err != nil {
			span.RecordError(err)
			slog.ErrorContext(ctx, "Failed to scan reply", "error", err)
			continue
		}
		replies = append(replies, reply)
	}

	span.SetAttributes(attribute.Int("replies.count", len(replies)))
	c.JSON(http.StatusOK, replies)
}
