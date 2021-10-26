package task

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go.uber.org/zap"
	"io"
	"sword-challenge/internal/user"
	"time"
)

type taskCrypto struct {
	gcm    cipher.AEAD
	logger *zap.SugaredLogger
	ivSize int
}

type encryptedTask struct {
	ID               int
	EncryptedSummary []byte     `db:"summary"`
	CompletedDate    *time.Time `db:"completed_date"`
	User             *user.User `db:"user"`
}

func NewCrypto(key string, logger *zap.SugaredLogger) (*taskCrypto, error) {
	hString, err := hex.DecodeString(key)
	if err != nil {
		logger.Warnw("Failed to decode key to hex binary")
		return nil, err
	}
	block, err := aes.NewCipher(hString)
	if err != nil {
		logger.Warnw("Failed to generate cypher")
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Warnw("Failed to generate cypher - step 2")
		return nil, err
	}

	return &taskCrypto{gcm: aesgcm, logger: logger, ivSize: 12}, nil
}

func (s *taskCrypto) encryptTask(t *task) (*encryptedTask, error) {
	nonce := make([]byte, s.ivSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		s.logger.Warnw("Failed to generate nonce")
		return nil, err
	}

	et := encryptedTask{ID: t.ID, CompletedDate: t.CompletedDate, User: t.User}
	et.EncryptedSummary = s.gcm.Seal(nonce, nonce, []byte(t.Summary), nil)
	return &et, nil
}

func (s *taskCrypto) decryptTask(et *encryptedTask, userId int) (*task, error) {
	s.logger.Infow("Task decryption requested", "taskId", et.ID, "userId", userId)

	if len(et.EncryptedSummary) < s.ivSize {
		s.logger.Warnw("Failed to parse nonce")
		return nil, fmt.Errorf("failed to parse nonce")
	}

	nonce, ciphertext := et.EncryptedSummary[:s.ivSize], et.EncryptedSummary[s.ivSize:]

	t := task{ID: et.ID, CompletedDate: et.CompletedDate, User: et.User}
	decryptedSummary, err := s.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		s.logger.Warnw("Failed to decrypt summary")
		return nil, err
	}

	t.Summary = string(decryptedSummary)
	return &t, nil
}
