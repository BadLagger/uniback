package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"uniback/dto"
	"uniback/repository"
	"uniback/utils"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	validate validator.Validate
	userRepo repository.UserRepository
}

func NewAuthController(u repository.UserRepository) *AuthController {
	return &AuthController{
		userRepo: u,
		validate: *validator.New(),
	}
}

func (c *AuthController) RegistrationHandler(w http.ResponseWriter, r *http.Request) {

	log := utils.GlobalLogger()

	log.Info("Get http reuest for registration from: %s", r.RemoteAddr)
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

	if err := c.validate.Struct(user); err != nil {
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
