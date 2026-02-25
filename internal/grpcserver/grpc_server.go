// Package grpcserver содержит реализацию gRPC-сервера сокращения URL.
package grpcserver

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/MarkelovSergey/url-shorter/internal/audit"
	"github.com/MarkelovSergey/url-shorter/internal/config"
	"github.com/MarkelovSergey/url-shorter/internal/middleware"
	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/service"
	"github.com/MarkelovSergey/url-shorter/internal/service/urlshorterservice"
	pb "github.com/MarkelovSergey/url-shorter/proto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	secretKey = "your-secret-key-change-in-production"
)

// ShortenerServer реализует gRPC-сервис ShortenerService.
type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer
	config            config.Config
	urlShorterService urlshorterservice.URLShorterService
	logger            *zap.Logger
	auditPublisher    audit.Publisher
}

// New создает новый экземпляр gRPC-сервера.
func New(
	cfg config.Config,
	urlShorterService urlshorterservice.URLShorterService,
	logger *zap.Logger,
	auditPublisher audit.Publisher,
) *ShortenerServer {
	return &ShortenerServer{
		config:            cfg,
		urlShorterService: urlShorterService,
		logger:            logger,
		auditPublisher:    auditPublisher,
	}
}

// ShortenURL обрабатывает запрос на создание короткой ссылки.
func (s *ShortenerServer) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	ctx = middleware.SetUserID(ctx, userID)

	reqURL := req.GetUrl()
	uParsed, parseErr := url.Parse(reqURL)
	if parseErr != nil || uParsed == nil ||
		(!strings.HasPrefix(reqURL, "http://") && !strings.HasPrefix(reqURL, "https://")) {
		return nil, status.Error(codes.InvalidArgument, "url not correct")
	}

	shortCode, genErr := s.urlShorterService.Generate(ctx, reqURL, userID)

	shortURL, joinErr := url.JoinPath(s.config.Server.BaseURL, shortCode)
	if joinErr != nil {
		return nil, status.Error(codes.Internal, "invalid URL format")
	}

	if genErr != nil {
		if errors.Is(genErr, service.ErrURLConflict) {
			s.auditPublisher.Publish(audit.NewEvent(audit.ActionShorten, reqURL, &userID))
			return &pb.URLShortenResponse{Result: shortURL}, status.Error(codes.AlreadyExists, "URL already shortened")
		}
		return nil, status.Error(codes.Internal, genErr.Error())
	}

	s.auditPublisher.Publish(audit.NewEvent(audit.ActionShorten, reqURL, &userID))

	return &pb.URLShortenResponse{Result: shortURL}, nil
}

// ExpandURL обрабатывает запрос на получение оригинального URL по короткому коду.
func (s *ShortenerServer) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	id := req.GetId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	originalURL, err := s.urlShorterService.GetOriginalURL(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrURLDeleted) {
			return nil, status.Error(codes.NotFound, "URL has been deleted")
		}
		return nil, status.Error(codes.NotFound, "ID not found")
	}

	s.auditPublisher.Publish(audit.NewEvent(audit.ActionFollow, originalURL, nil))

	return &pb.URLExpandResponse{Result: originalURL}, nil
}

// ListUserURLs обрабатывает запрос на получение списка URL пользователя.
func (s *ShortenerServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	records, getErr := s.urlShorterService.GetUserURLs(ctx, userID)
	if getErr != nil {
		s.logger.Error("Failed to get user URLs", zap.Error(getErr))
		return nil, status.Error(codes.Internal, "failed to get user URLs")
	}

	urls := make([]*pb.URLData, 0, len(records))
	for _, record := range records {
		shortURL, joinErr := url.JoinPath(s.config.Server.BaseURL, record.ShortURL)
		if joinErr != nil {
			s.logger.Error("Failed to join URL", zap.Error(joinErr))
			continue
		}
		urls = append(urls, &pb.URLData{
			ShortUrl:    shortURL,
			OriginalUrl: record.OriginalURL,
		})
	}

	return &pb.UserURLsResponse{Url: urls}, nil
}

// AuthInterceptor возвращает gRPC UnaryServerInterceptor для аутентификации.
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		var userID string

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			values := md.Get("authorization")
			if len(values) > 0 {
				token := values[0]
				if id, valid := validateJWT(token); valid {
					userID = id
				}
			}
		}

		if userID == "" {
			userID = uuid.New().String()
		}

		ctx = middleware.SetUserID(ctx, userID)

		// Генерируем токен и добавляем в response metadata
		tokenString, err := generateJWT(userID)
		if err == nil {
			header := metadata.Pairs("authorization", tokenString)
			_ = grpc.SetHeader(ctx, header)
		}

		return handler(ctx, req)
	}
}

// getUserID извлекает ID пользователя из контекста gRPC.
func (s *ShortenerServer) getUserID(ctx context.Context) (string, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok || userID == "" {
		return "", status.Error(codes.Unauthenticated, "user not authenticated")
	}
	return userID, nil
}

func validateJWT(tokenString string) (string, bool) {
	token, err := jwt.ParseWithClaims(tokenString, &model.UserClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return "", false
	}

	if claims, ok := token.Claims.(*model.UserClaims); ok && token.Valid {
		return claims.UserID, true
	}

	return "", false
}

func generateJWT(userID string) (string, error) {
	claims := model.UserClaims{
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}
