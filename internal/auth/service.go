package auth

import (
	"context"
	"errors"
	"fmt"
	"gw-currency-wallet/internal/cache"
	"gw-currency-wallet/internal/proto/proto/exchange"
	"gw-currency-wallet/internal/storages"
	"gw-currency-wallet/pkg/logging"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	storage         storages.Repository
	jwtSecret       string
	rateCache       *cache.RateCache
	logger          logging.Logger
	exchangerClient exchange.ExchangeServiceClient
}

func NewService(storage storages.Repository, jwtSecret string, exClient exchange.ExchangeServiceClient, logger *logging.Logger) *Service {
	return &Service{
		storage:         storage,
		jwtSecret:       jwtSecret,
		rateCache:       cache.NewRateCache(30 * time.Second),
		logger:          *logger,
		exchangerClient: exClient,
	}
}
func (s *Service) Register(ctx context.Context, email, password string) error {
	passwordHash := hashPassword(password)
	_, err := s.storage.CreateUser(ctx, email, passwordHash)
	return err
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.storage.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("user not found")
	}

	if !checkPassword(password, user.PasswordHash) {
		return "", errors.New("invalid password")
	}

	return s.generateToken(user.ID)
}

func (s *Service) ParseToken(tokenStr string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return 0, errors.New("invalid token")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, errors.New("invalid token claims")
}

func (s *Service) GetExchangeRateWithCache(from, to string) (float32, error) {
	// Сначала пробуем кэш
	if rate, ok := s.rateCache.GetRate(from, to); ok {
		s.logger.Infof("Get Rate from cache %v", rate)
		return rate, nil
	}

	resp, err := s.exchangerClient.GetExchangeRateForCurrency(context.Background(), &exchange.CurrencyRequest{
		FromCurrency: from,
		ToCurrency:   to,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	s.logger.Infof("Get Rate from exchangerClient %v", resp.Rate)
	return resp.Rate, nil
}

// FetchAndCacheAllRates — вызывается при /exchange/rates
func (s *Service) FetchAndCacheAllRates() (map[string]float32, error) {
	resp, err := s.exchangerClient.GetExchangeRates(context.Background(), &exchange.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all rates: %w", err)
	}

	// Сохраняем в кэш
	s.rateCache.SetAllRates(resp.Rates)
	return resp.Rates, nil
}

func (s *Service) generateToken(userId int64) (string, error) {
	claims := Claims{
		UserID: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
