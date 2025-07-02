package service

import (
	"crypto"
	"crypto/rand"
	"os"
	"time"
	"uniback/utils"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

type PgpHmacConfig struct {
	pgpPublicFile  string
	pgpPrivateFile string
	hmacKey        string
}

type PgpHmacService struct {
	cfg           PgpHmacConfig
	pgpPublicKey  openpgp.Entity
	pgpPrivateKey openpgp.Entity
}

func PgpHmacConfgiFromGlobalConfig(cfg *utils.Config) *PgpHmacConfig {
	return &PgpHmacConfig{
		pgpPublicFile:  cfg.PgpPublicPath,
		pgpPrivateFile: cfg.PgpPrivatePath,
		hmacKey:        cfg.HmacKey,
	}
}

func NewPgpHmacService(cfg *PgpHmacConfig) *PgpHmacService {
	log := utils.GlobalLogger()

	log.Info("Create PGP HMAC Service")

	needCreateKeys := false
	if _, err := os.Stat(cfg.pgpPublicFile); err != nil {
		log.Info("No public key file: %s", cfg.pgpPublicFile)
		if _, err := os.Stat(cfg.pgpPrivateFile); err == nil {
			log.Info("Delete private key file: %s", cfg.pgpPrivateFile)
			os.Remove(cfg.pgpPrivateFile)
		}
		needCreateKeys = true
	}

	if !needCreateKeys {
		if _, err := os.Stat(cfg.pgpPrivateFile); err != nil {
			log.Info("No private key file: %s", cfg.pgpPrivateFile)
			log.Info("Delete public key file: %s", cfg.pgpPublicFile)
			os.Remove(cfg.pgpPublicFile)
			needCreateKeys = true
		}
	}

	var newService PgpHmacService

	newService.cfg = *cfg

	if needCreateKeys {
		err := newService.createNewPgp()
		if err != nil {
			log.Critical("Can't create pgp")
			return nil
		}
	}

	key, err := newService.readKey(cfg.pgpPublicFile)
	if err != nil {
		log.Critical("Can't read public key: %w", err)
		return nil
	}

	newService.pgpPublicKey = *key

	key, err = newService.readKey(cfg.pgpPrivateFile)
	if err != nil {
		log.Critical("Can't read private key: %w", err)
		return nil
	}

	newService.pgpPrivateKey = *key

	return &newService
}

func (s *PgpHmacService) createNewPgp() error {

	log := utils.GlobalLogger()

	log.Info("Try to create new PGP")
	keyConfig := &packet.Config{
		DefaultHash:   crypto.SHA256,        // Алгоритм хеширования
		DefaultCipher: packet.CipherAES256,  // Шифрование приватного ключа
		RSABits:       4096,                 // Размер RSA-ключа
		Algorithm:     packet.PubKeyAlgoRSA, // Алгоритм ключа
		Time:          func() time.Time { return time.Now() },
		Rand:          rand.Reader,
	}

	entity, err := openpgp.NewEntity(
		"Uniback",           // Имя
		"",                  // Комментарий (опционально)
		"email@example.com", // Email
		keyConfig,
	)
	if err != nil {
		log.Critical("Can't create new pgp entety: %w", err)
		return err
	}

	pubKeyFile, err := os.Create(s.cfg.pgpPublicFile)
	if err != nil {
		log.Critical("Can't public key file: %w", err)
		return err
	}
	defer pubKeyFile.Close()

	pubKeyWriter, err := armor.Encode(pubKeyFile, openpgp.PublicKeyType, nil)
	if err != nil {
		log.Critical("Can't encode pgp armor: %w", err)
		return err
	}
	defer pubKeyWriter.Close()

	err = entity.Serialize(pubKeyWriter)
	if err != nil {
		log.Critical("Can't serialized pgp: %w", err)
		return err
	}

	privKeyFile, err := os.Create(s.cfg.pgpPrivateFile)
	if err != nil {
		log.Critical("Can't private key file: %w", err)
		return err
	}
	defer privKeyFile.Close()

	privKeyWriter, err := armor.Encode(privKeyFile, openpgp.PrivateKeyType, nil)
	if err != nil {
		log.Critical("Can't encode pgp armor: %w", err)
		return err
	}
	defer privKeyWriter.Close()

	err = entity.SerializePrivate(privKeyWriter, keyConfig)
	if err != nil {
		log.Critical("Can't serialized pgp: %w", err)
		return err
	}

	log.Info("New PGP created OK!!!")

	return nil
}

func (cs *PgpHmacService) PgpEncode(data string) []byte {
	return nil
}

func (cs *PgpHmacService) PgpDecode(data string) []byte {
	return nil
}

func (s *PgpHmacService) readKey(fPath string) (*openpgp.Entity, error) {
	file, err := os.Open(fPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	block, err := armor.Decode(file)
	if err != nil {
		return nil, err
	}

	keyData := packet.NewReader(block.Body)

	entity, err := openpgp.ReadEntity(keyData)
	if err != nil {
		return nil, err
	}
	return entity, nil
}
