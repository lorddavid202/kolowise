package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB        *pgxpool.Pool
	JWTSecret string
	JWTIssuer string
}

func NewHandler(db *pgxpool.Pool, jwtSecret, jwtIssuer string) *Handler {
	return &Handler{
		DB:        db,
		JWTSecret: jwtSecret,
		JWTIssuer: jwtIssuer,
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	query := `
		INSERT INTO users (full_name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, full_name, email, currency, created_at
	`

	var user UserResponse
	err = h.DB.QueryRow(
		c.Request.Context(),
		query,
		req.FullName,
		req.Email,
		hashedPassword,
	).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Currency,
		&user.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	token, err := GenerateToken(h.JWTSecret, h.JWTIssuer, user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		User:        user,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	query := `
		SELECT id, full_name, email, currency, password_hash, created_at
		FROM users
		WHERE LOWER(email) = LOWER($1) AND is_active = TRUE
		LIMIT 1
	`

	var user UserResponse
	var passwordHash string

	err := h.DB.QueryRow(
		c.Request.Context(),
		query,
		req.Email,
	).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Currency,
		&passwordHash,
		&user.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if err := ComparePassword(passwordHash, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, err := GenerateToken(h.JWTSecret, h.JWTIssuer, user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	_, _ = h.DB.Exec(
		c.Request.Context(),
		`UPDATE users SET updated_at = $1 WHERE id = $2`,
		time.Now().UTC(),
		user.ID,
	)

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		User:        user,
	})
}

func (h *Handler) Me(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	query := `
		SELECT id, full_name, email, currency, created_at
		FROM users
		WHERE id = $1 AND is_active = TRUE
		LIMIT 1
	`

	var user UserResponse
	err := h.DB.QueryRow(
		c.Request.Context(),
		query,
		userIDVal.(string),
	).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Currency,
		&user.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
