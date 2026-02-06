package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"why-backend/internal/auth"
	"why-backend/internal/config"
	"why-backend/internal/models"
)

var authTracer = otel.Tracer("why-backend/handlers/auth")

type AuthHandler struct {
	db     *sql.DB
	config *config.Config
}

func NewAuthHandler(db *sql.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		db:     db,
		config: cfg,
	}
}

// Signup creates a new user account
func (h *AuthHandler) Signup(c *gin.Context) {
	ctx, span := authTracer.Start(c.Request.Context(), "Signup")
	defer span.End()

	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	span.SetAttributes(attribute.String("user.email", req.Email))

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to hash password", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Create user
	var user models.User
	err = h.db.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2)
		 RETURNING id, email, created_at, updated_at`,
		req.Email, passwordHash,
	).Scan(&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to create user", "error", err, "email", req.Email)
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email, h.config.JWTSecret)
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to generate token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	span.SetAttributes(attribute.String("user.id", user.ID))
	slog.InfoContext(ctx, "User created successfully", "user_id", user.ID, "email", user.Email)

	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login authenticates a user and returns a JWT token
func (h *AuthHandler) Login(c *gin.Context) {
	ctx, span := authTracer.Start(c.Request.Context(), "Login")
	defer span.End()

	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	span.SetAttributes(attribute.String("user.email", req.Email))

	// Get user by email
	var user models.User
	err := h.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = $1`,
		req.Email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		span.SetAttributes(attribute.Bool("auth.failed", true))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	} else if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Database error during login", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	// Check password
	if err := auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		span.SetAttributes(attribute.Bool("auth.failed", true))
		slog.WarnContext(ctx, "Failed login attempt", "email", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email, h.config.JWTSecret)
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to generate token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	span.SetAttributes(
		attribute.String("user.id", user.ID),
		attribute.Bool("auth.success", true),
	)
	slog.InfoContext(ctx, "User logged in successfully", "user_id", user.ID, "email", user.Email)

	c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	})
}
