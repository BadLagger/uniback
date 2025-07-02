package service

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
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
	var encryptedBuf bytes.Buffer

	armorWriter, _ := armor.Encode(&encryptedBuf, "PGP MESSAGE", nil)
	encryptWriter, _ := openpgp.Encrypt(armorWriter, []*openpgp.Entity{&cs.pgpPublicKey}, nil, nil, nil)

	encryptWriter.Write([]byte(data))
	encryptWriter.Close()
	armorWriter.Close()

	return encryptedBuf.Bytes()
}

func (cs *PgpHmacService) PgpDecode(data []byte) string {
	encryptedBuf := bytes.NewBuffer(data)

	block, _ := armor.Decode(bytes.NewReader(encryptedBuf.Bytes()))
	keyRing := &openpgp.EntityList{&cs.pgpPrivateKey}
	md, _ := openpgp.ReadMessage(block.Body, keyRing, nil, nil)
	decrypted, _ := io.ReadAll(md.UnverifiedBody)
	return string(decrypted)
}

func (cs *PgpHmacService) GenerateCardLuhn() (string, error) {
	prefix := "2" // Карты "Мир" начинаются с 2
	length := 16  // Стандартная длина номера карты

	// Генерируем случайные цифры (кроме последней)
	randomPartLength := length - len(prefix) - 1
	randomPart := ""
	for i := 0; i < randomPartLength; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("Random number gen error: %v", err)
		}
		randomPart += num.String()
	}

	// Собираем номер без последней цифры
	partialNumber := prefix + randomPart

	// Вычисляем контрольную цифру по алгоритму Луна
	checkDigit := calculateLuhnCheckDigit(partialNumber)

	// Возвращаем полный номер карты
	return partialNumber + strconv.Itoa(checkDigit), nil
}

func (cs *PgpHmacService) GenerateCvv() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000))
	return fmt.Sprintf("%03d", n)
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

func calculateLuhnCheckDigit(number string) int {
	sum := 0

	for i := 0; i < len(number); i++ {
		digit, _ := strconv.Atoi(string(number[len(number)-1-i]))

		if i%2 == 0 {
			digit *= 2
			if digit > 9 {
				digit = digit/10 + digit%10
			}
		}

		sum += digit
	}

	return (10 - (sum % 10)) % 10
}
