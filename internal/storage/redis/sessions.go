package redis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage"
)

type SessionStorage struct {
	client *redis.Client
	logger *slog.Logger
}

func NewSessionStorage(client *redis.Client) *SessionStorage {
	return &SessionStorage{
		client: client,
		logger: slog.Default().With("struct", "redis.SessionStorage"),
	}
}

func (s *SessionStorage) encode(in *entity.SessionInfo) (io.Reader, error) {
	buf := new(bytes.Buffer)

	if err := json.NewEncoder(buf).Encode(in); err != nil {
		return nil, err
	}

	return buf, nil
}

func (s *SessionStorage) decode(in io.Reader) (*entity.SessionInfo, error) {
	info := new(entity.SessionInfo)

	if err := json.NewDecoder(in).Decode(info); err != nil {
		return nil, err
	}

	return info, nil
}

func (s *SessionStorage) Save(ctx context.Context, in *entity.SessionInfo) (string, error) {
	log := s.logger.With(slog.String("method", "Save"))

	buf, err := s.encode(in)
	if err != nil {
		log.Error("error encoding session", sl.Err(err), slog.Any("in", in))
		return "", nil
	}

	id := uuid.New().String()

	content, err := io.ReadAll(buf)
	if err != nil {
		log.Error("error reading session", sl.Err(err))
		return "", nil
	}

	if _, err := s.client.Set(ctx, id, content, 0).Result(); err != nil {
		log.Error("error saving session", sl.Err(err))
		return "", nil
	}

	return id, nil
}

func (s *SessionStorage) Find(ctx context.Context, id string) (*entity.SessionInfo, error) {
	log := s.logger.With(slog.String("method", "Find"))

	val, err := s.client.Get(ctx, id).Result()
	if errors.Is(err, redis.Nil) {
		log.Debug("session not found", slog.String("id", id))
		return nil, storage.ErrSessionNotFound
	}
	if err != nil {
		log.Error("error getting session", slog.String("id", id), sl.Err(err))
		return nil, err
	}

	info, err := s.decode(strings.NewReader(val))
	if err != nil {
		log.Error(
			"error decoding session",
			slog.String("id", id),
			slog.String("val", val),
			sl.Err(err),
		)
		return nil, err
	}

	return info, nil
}
