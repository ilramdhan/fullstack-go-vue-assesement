package usecase

import (
	"fmt"
	"time"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	"github.com/durianpay/fullstack-boilerplate/internal/module/auth/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID string
	Email  string
	Role   string
}

type AuthUsecase interface {
	Login(email string, password string) (string, *entity.User, error)
	Verify(token string) (*Claims, error)
}

type Auth struct {
	repo      repository.UserRepository
	jwtSecret []byte
	ttl       time.Duration
}

func NewAuthUsecase(repo repository.UserRepository, jwtSecret []byte, ttl time.Duration) *Auth {
	return &Auth{repo: repo, jwtSecret: jwtSecret, ttl: ttl}
}

func (a *Auth) Login(email string, password string) (string, *entity.User, error) {
	user, err := a.repo.GetUserByEmail(email)
	if err != nil {
		return "", nil, err
	}
	if user.ID == "" {
		return "", nil, entity.ErrorNotFound("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, entity.ErrorUnauthorized("invalid credentials")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"role":  user.Role,
		"iat":   now.Unix(),
		"exp":   now.Add(a.ttl).Unix(),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.jwtSecret)
	if err != nil {
		return "", nil, entity.WrapError(err, entity.ErrorCodeInternal, "sign token")
	}
	return signed, user, nil
}

func (a *Auth) Verify(token string) (*Claims, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Method.Alg())
		}
		return a.jwtSecret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, entity.ErrorUnauthorized("invalid or expired token")
	}

	mc, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, entity.ErrorUnauthorized("invalid token claims")
	}

	c := &Claims{}
	if v, ok := mc["sub"].(string); ok {
		c.UserID = v
	}
	if v, ok := mc["email"].(string); ok {
		c.Email = v
	}
	if v, ok := mc["role"].(string); ok {
		c.Role = v
	}
	if c.UserID == "" || c.Role == "" {
		return nil, entity.ErrorUnauthorized("incomplete token claims")
	}
	return c, nil
}
