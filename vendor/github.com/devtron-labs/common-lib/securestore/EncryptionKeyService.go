/*
 * Copyright (c) 2024. Devtron Inc.
 */

package securestore

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/go-pg/pg"
	log "github.com/sirupsen/logrus"
)

func SetEncryptionKey() error {
	repo, err := NewAttributesRepositoryImplForDatabase("orchestrator") //hardcoded for orchestrator as need to pick a common key for every service
	if err != nil {
		log.Println("error in creating attributes repository", "err", err)
		return err
	}
	encryptionService := NewEncryptionKeyServiceImpl(repo)
	err = encryptionService.CreateAndStoreEncryptionKey()
	if err != nil {
		log.Println("error in creating and storing encryption key", "err", err)
		return err
	}
	return nil
}

type EncryptionKeyService interface {
	// CreateAndStoreEncryptionKey generates a new AES-256 encryption key and stores it in the attributes repository
	CreateAndStoreEncryptionKey() error

	// RotateEncryptionKey generates a new encryption key and stores it (deactivating the old one)
	RotateEncryptionKey(userId int32) (string, error)

	// GenerateEncryptionKey generates a new AES-256 encryption key (32 bytes)
	GenerateEncryptionKey() (string, error)

	GetEncryptionKey() (string, error)
}

type EncryptionKeyServiceImpl struct {
	attributesRepository AttributesRepository
}

func NewEncryptionKeyServiceImpl(attributesRepository AttributesRepository) *EncryptionKeyServiceImpl {
	impl := &EncryptionKeyServiceImpl{
		attributesRepository: attributesRepository,
	}
	return impl
}

// GenerateEncryptionKey generates a new AES-256 encryption key (32 bytes = 256 bits)
func (impl *EncryptionKeyServiceImpl) GenerateEncryptionKey() (string, error) {
	// Generate 32 random bytes for AES-256
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		log.Error("error generating random encryption key", "err", err)
		return "", fmt.Errorf("failed to generate encryption key: %w", err)
	}
	// Encode to hex string for storage
	keyHex := hex.EncodeToString(key)
	return keyHex, nil
}

// CreateAndStoreEncryptionKey generates a new AES-256 encryption key and stores it in the attributes repository
func (impl *EncryptionKeyServiceImpl) CreateAndStoreEncryptionKey() error {
	// Check if encryption key already exists
	encryptionKeyModel, err := impl.attributesRepository.FindByKey(ENCRYPTION_KEY)
	if err != nil && err != pg.ErrNoRows {
		log.Error("error checking for existing encryption key", "err", err)
		return err
	}
	var encryptionKeyEncoded string
	if encryptionKeyModel != nil && encryptionKeyModel.Id > 0 && len(encryptionKeyModel.Value) > 0 {
		encryptionKeyEncoded = encryptionKeyModel.Value
		log.Println("encryption key already exists", "keyId", encryptionKeyModel.Id)
	} else {
		// Generate new encryption key
		encryptionKeyNew, err := impl.GenerateEncryptionKey()
		if err != nil {
			return err
		}
		// Store in repository
		err = impl.attributesRepository.SaveEncryptionKeyIfNotExists(encryptionKeyNew)
		if err != nil {
			log.Error("error storing encryption key", "err", err)
			return fmt.Errorf("failed to store encryption key: %w", err)
		}
		encryptionKeyEncoded = encryptionKeyNew
		log.Println("Successfully created and stored encryption key")
	}

	encryptionKey, err = hex.DecodeString(encryptionKeyEncoded)
	if err != nil || len(encryptionKey) != 32 {
		return fmt.Errorf("encryptionKey is incorrect : %v", err)
	}
	return nil
}

// RotateEncryptionKey generates a new encryption key and stores it (deactivating the old one)
func (impl *EncryptionKeyServiceImpl) RotateEncryptionKey(userId int32) (string, error) {
	log.Println("Rotating encryption key", "userId", userId)

	// Generate new encryption key
	newEncryptionKey, err := impl.GenerateEncryptionKey()
	if err != nil {
		return "", err
	}

	// Store in repository (this will deactivate the old key)
	err = impl.attributesRepository.SaveEncryptionKeyIfNotExists(newEncryptionKey)
	if err != nil {
		log.Error("error rotating encryption key", "err", err)
		return "", fmt.Errorf("failed to rotate encryption key: %w", err)
	}
	//TODO: also need to rotate encryption's already done
	log.Println("Successfully rotated encryption key", "userId", userId)
	return newEncryptionKey, nil
}

// GetEncryptionKey retrieves the active encryption key from the repository
func (impl *EncryptionKeyServiceImpl) GetEncryptionKey() (string, error) {
	key, err := impl.attributesRepository.GetEncryptionKey()
	if err != nil {
		if err == pg.ErrNoRows {
			log.Error("encryption key not found in repository")
			return "", fmt.Errorf("encryption key not found, please create one first")
		}
		log.Error("error retrieving encryption key", "err", err)
		return "", err
	}
	return key, nil
}
