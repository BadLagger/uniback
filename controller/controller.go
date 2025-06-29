package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"uniback/dto"
	"uniback/repository"
	"uniback/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type JWTClaims struct {
	Username string
	UserID   int
	jwt.StandardClaims
}

type AuthController struct {
	validate  validator.Validate
	userRepo  repository.UserRepository
	secretKey string
}

func NewAuthController(u repository.UserRepository, s string) *AuthController {
	return &AuthController{
		userRepo:  u,
		validate:  *validator.New(),
		secretKey: s,
	}
}

func (c *AuthController) RegistrationHandler(w http.ResponseWriter, r *http.Request) {

	log := utils.GlobalLogger()

	log.Info("Get http request for registration from: %s", r.RemoteAddr)
	// !todo создать более подробный лог источника запроса (разобрать headers)

	if r.Method != http.MethodPost {
		log.Error("Wrong method!")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user dto.UserCreateRequest

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Error("Json parse error: %w", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = c.validateRequest(w, user); err != nil {
		return
	}

	userExists, emailExists, phoneExists, err := c.userRepo.IsUserExistsByUsernameEmailPhone(r.Context(), user)
	if err != nil {
		log.Critical("Check user in DB error: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if userExists {
		log.Error("User %s exists", user.Username)
		http.Error(w, "User with the same username exists already", http.StatusConflict)
		return
	}

	if emailExists {
		log.Error("User %s try to register with already registred email %s", user.Username, user.Email)
		http.Error(w, "User with the same email exists already", http.StatusConflict)
		return
	}

	if phoneExists {
		log.Error("User %s try to register with already registred phone %s", user.Username, user.Phone)
		http.Error(w, "User with the same phone exists already", http.StatusConflict)
		return
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		log.Critical("Try to make hash error: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Debug("Hash for user: %s", string(hashPassword))
	user.Password = string(hashPassword)

	err = c.userRepo.CreateUser(r.Context(), user)
	if err != nil {
		log.Critical("Can't to write user %s to db: %w", user.Username, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info("User %s created!", user.Username)
	w.WriteHeader(http.StatusOK)
}

func (c *AuthController) LoginHandler(w http.ResponseWriter, r *http.Request) {

	log := utils.GlobalLogger()

	log.Info("Get http request for login from: %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		log.Error("Wrong method!")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user dto.UserLoginRequest

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Error("Json parse error: %w", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = c.validateRequest(w, user); err != nil {
		return
	}

	userFromDb, err := c.userRepo.GetUserByUsername(r.Context(), user.Username)

	if err != nil {
		log.Error("Getting %s from db error: %w", user.Username, err)
		http.Error(w, "Wrong user login", http.StatusForbidden)
		return
	}

	log.Debug("Get user from db: %s, %s", userFromDb.Name, userFromDb.Password)

	err = bcrypt.CompareHashAndPassword([]byte(userFromDb.Password), []byte(user.Password))
	if err != nil {
		log.Error("Invalid password for %s (%w)", user.Username, err)
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		Username: userFromDb.Name,
		UserID:   userFromDb.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	})

	tokenStr, err := token.SignedString([]byte(c.secretKey))

	if err != nil {
		log.Critical("Failed to generate jwt: %w", err)
		http.Error(w, "Failed to generate jwt", http.StatusInternalServerError)
		return
	}

	log.Debug("For user %s jwt: %s", user.Username, tokenStr)

	w.Header().Set("Authorization", "Bearer "+tokenStr)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"message": "successful",
		"user":    user.Username,
	})
}

func (c *AuthController) validateRequest(w http.ResponseWriter, s interface{}) error {

	log := utils.GlobalLogger()

	if err := c.validate.Struct(s); err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			log.Error("Field %s failed validator (%s = %s)",
				err.Field(),
				err.Tag(),
				err.Param())
			validationErrors = append(validationErrors, fmt.Sprintf(
				"Field %s failed validator (%s = %s)",
				err.Field(),
				err.Tag(),
				err.Param(),
			))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error":   "Validation failed",
			"details": validationErrors,
		})
		return err
	}
	return nil
}
