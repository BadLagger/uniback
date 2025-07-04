package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"uniback/dto"
	"uniback/models"
	"uniback/repository"
	"uniback/service"
	"uniback/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type JWTClaims struct {
	Username string
	jwt.StandardClaims
}

type AuthController struct {
	validate      validator.Validate
	userRepo      repository.UserRepository
	service       service.Service
	cryptoService service.CryptoService
	secretKey     string
}

func NewAuthController(u repository.UserRepository, cs service.CryptoService, sr service.Service, s string) *AuthController {
	return &AuthController{
		userRepo:      u,
		service:       sr,
		cryptoService: cs,
		validate:      *validator.New(),
		secretKey:     s,
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

func (c *AuthController) AccountsHandler(w http.ResponseWriter, r *http.Request) {
	log := utils.GlobalLogger()

	log.Info("Get http request for Account from: %s", r.RemoteAddr)

	if r.Method != http.MethodGet {
		log.Error("Wrong method!")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value("jwtClaims").(*JWTClaims)
	if !ok {
		log.Critical("No jwt claims in context")
		http.Error(w, "Failed to get claims", http.StatusInternalServerError)
		return
	}

	log.Info("Username %s request for accounts", claims.Username)

	accounts, err := c.userRepo.GetAccountsByUsername(r.Context(), claims.Username)
	if err != nil {
		log.Critical("DB error: %w", err)
		http.Error(w, "Failed to get accounts from DB", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(accounts)
	if err != nil {
		log.Critical("Encode accounts to json error: %w", err)
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", r.Header.Get("Authorization"))
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (c *AuthController) AccountsCreateHandler(w http.ResponseWriter, r *http.Request) {
	log := utils.GlobalLogger()
	log.Info("Get http request for Createnig Account from: %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		log.Error("Wrong method!")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value("jwtClaims").(*JWTClaims)
	if !ok {
		log.Critical("No jwt claims in context")
		http.Error(w, "Failed to get claims", http.StatusInternalServerError)
		return
	}

	var accountRequest dto.AccountCreateRequestDto

	err := json.NewDecoder(r.Body).Decode(&accountRequest)
	if err != nil {
		log.Error("Json parse error: %w", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = c.validateRequest(w, accountRequest); err != nil {
		return
	}

	userId, err := c.userRepo.GetUserId(r.Context(), claims.Username)

	if err != nil {
		log.Error("Can't get user id from DB: %w", err)
		http.Error(w, "Failed to get user id from DB", http.StatusBadRequest)
		return
	}

	newAccountNumber := models.GenerateAccount()
	for {
		isExists, err := c.userRepo.IsAccountExits(r.Context(), newAccountNumber)

		if err != nil {
			log.Critical("Can't check account existence in DB: %w", err)
			http.Error(w, "Can't check account existence in DB", http.StatusInternalServerError)
			return
		}

		if !isExists {
			break
		}
		newAccountNumber = models.GenerateAccount()
	}

	log.Debug("Cenerate new number: %s", newAccountNumber)

	responseDto, err := c.userRepo.CreateAccount(r.Context(), models.Account{
		UserId:        userId,
		AccountNumber: newAccountNumber,
		AccountType:   accountRequest.AccountType,
		Status:        "active",
	})

	if err != nil {
		log.Error("Create Account error: %w", err)
		http.Error(w, "Can't create account!", http.StatusBadRequest)
	}

	jsonData, err := json.Marshal(responseDto)
	if err != nil {
		log.Critical("Encode accounts to json error: %w", err)
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", r.Header.Get("Authorization"))
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (c *AuthController) DepositHandler(w http.ResponseWriter, r *http.Request) {
	c.transactionRequest(w, r, c.service.DepositTransaction)
}

func (c *AuthController) WithdrawalHandler(w http.ResponseWriter, r *http.Request) {
	c.transactionRequest(w, r, c.service.WithdrawalTransaction)
}

func (c *AuthController) TransferHandler(w http.ResponseWriter, r *http.Request) {
	log := utils.GlobalLogger()
	log.Info("Get http request for Transaction Account from: %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		log.Error("Wrong method!")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value("jwtClaims").(*JWTClaims)
	if !ok {
		log.Critical("No jwt claims in context")
		http.Error(w, "Failed to get claims", http.StatusInternalServerError)
		return
	}

	var transferDto dto.TransferRequestDto

	err := json.NewDecoder(r.Body).Decode(&transferDto)
	if err != nil {
		log.Error("Json parse error: %w", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validateRequest(w, transferDto); err != nil {
		return
	}

	sourceAccount, err := c.userRepo.GetAccountByUsername(r.Context(), transferDto.SourceAccountNumber, claims.Username)

	if err != nil {
		log.Error("Error confirm account: %w", err)
		http.Error(w, "Wrong source account number", http.StatusBadRequest)
		return
	}

	if sourceAccount.Status != "active" {
		log.Error("Try to perform trasaction with not active account")
		http.Error(w, "Try to perform trasaction with not active account", http.StatusBadRequest)
		return
	}

	destAccount, err := c.userRepo.GetAccountByNumber(r.Context(), transferDto.DestinationAccountNumber)

	if err != nil {
		log.Error("Error destination account: %w", err)
		http.Error(w, "Wrong destination account number", http.StatusBadRequest)
		return
	}

	if destAccount.Status != "active" {
		log.Error("Try to perform trasaction with not active account")
		http.Error(w, "Try to perform trasaction with not active account", http.StatusBadRequest)
		return
	}

	account, err := c.service.TransferTransaction(r.Context(), *sourceAccount, *destAccount, transferDto.Amount)

	if err != nil {
		log.Error("transaction error: %w", err)
		http.Error(w, "transaction error", http.StatusBadRequest)
		return
	}

	jsonData, err := json.Marshal(dto.AccountToAccountReponseDto(account))
	if err != nil {
		log.Critical("Encode accounts to json error: %w", err)
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", r.Header.Get("Authorization"))
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (c *AuthController) NewCardHandler(w http.ResponseWriter, r *http.Request) {
	log := utils.GlobalLogger()
	log.Info("Get http request for Create New Card from: %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		log.Error("Wrong method!")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value("jwtClaims").(*JWTClaims)
	if !ok {
		log.Critical("No jwt claims in context")
		http.Error(w, "Failed to get claims", http.StatusInternalServerError)
		return
	}

	var newCardRequest dto.CardCreateRequestDto

	err := json.NewDecoder(r.Body).Decode(&newCardRequest)
	if err != nil {
		log.Error("Json parse error: %w", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validateRequest(w, newCardRequest); err != nil {
		return
	}

	cardAccount, err := c.userRepo.GetAccountByUsername(r.Context(), newCardRequest.AccountNumber, claims.Username)

	if err != nil {
		log.Error("Error confirm account: %w", err)
		http.Error(w, "Wrong source account number", http.StatusBadRequest)
		return
	}

	userId, err := c.userRepo.GetUserId(r.Context(), claims.Username)
	if err != nil {
		log.Error("Error getting userid: %w", err)
		http.Error(w, "Bad user ID", http.StatusBadRequest)
		return
	}

	log.Debug("Card account: %s", cardAccount.AccountNumber)
	log.Debug("Card account id: %d", cardAccount.Id)
	log.Debug("UserId: %d", userId)

	var newLuhnNumber string
	var secureLuhnNumber []byte

	for {
		newLuhnNumber, err = c.cryptoService.GenerateCardLuhn()
		if err != nil {
			log.Error("Error generate Luhn number: %w", err)
			http.Error(w, "Error generate Luhn number", http.StatusInternalServerError)
			return
		}
		secureLuhnNumber = c.cryptoService.PgpEncode(newLuhnNumber)

		isExist, err := c.userRepo.IsCardExists(r.Context(), secureLuhnNumber)
		if err != nil {
			log.Error("Can't read cards from DB: %w", err)
			return
		}

		if !isExist {
			break
		}
	}

	currentTime := time.Now()
	expiredTime := time.Now().AddDate(5, 0, 0)
	secureExpiredTime := c.cryptoService.PgpEncode(expiredTime.String())

	newCvv := c.cryptoService.GenerateCvv()
	secureCvv := c.cryptoService.PgpEncode(newCvv)

	err = c.userRepo.CreateNewCard(r.Context(), models.Card{
		AccountId: cardAccount.Id,
		Number:    secureLuhnNumber,
		Expiry:    secureExpiredTime,
		Cvv:       secureCvv,
		CreatedAt: currentTime,
	})

	if err != nil {
		log.Error("Can't save card: %w", err)
		http.Error(w, "Cant save card in DB", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", r.Header.Get("Authorization"))
	w.WriteHeader(http.StatusOK)
}

func (ac *AuthController) ShowCardsHandler(w http.ResponseWriter, r *http.Request) {

}

func (ac *AuthController) ShowCreditsHanlder(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Service not implemented yet", http.StatusNotImplemented)
}

func (ac *AuthController) NewCreditHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Service not implemented yet", http.StatusNotImplemented)
}

func (ac *AuthController) AnalyticsHanlder(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Service not implemented yet", http.StatusNotImplemented)
}

func (ac *AuthController) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := utils.GlobalLogger()
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Error("No autorization token")
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[len("Bearer "):]
		claims := &JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(ac.secretKey), nil
		})

		log.Debug("Username from token: %s", claims.Username)

		if err != nil || !token.Valid {
			log.Error("Invalid token")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		isUser, err := ac.userRepo.IsUserExists(r.Context(), claims.Username)

		if err != nil {
			log.Critical("DB Error: %w", err)
			http.Error(w, "DB Error", http.StatusInternalServerError)
			return
		}

		if !isUser {
			log.Error("Wrong token from user %s", claims.Username)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "jwtClaims", claims)

		next(w, r.WithContext(ctx))
	}
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

func (c *AuthController) transactionRequest(w http.ResponseWriter, r *http.Request, transaction func(ctx context.Context, acc models.Account, amount float64) (*models.Account, error)) {
	log := utils.GlobalLogger()
	log.Info("Get http request for Transaction Account from: %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		log.Error("Wrong method!")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value("jwtClaims").(*JWTClaims)
	if !ok {
		log.Critical("No jwt claims in context")
		http.Error(w, "Failed to get claims", http.StatusInternalServerError)
		return
	}

	var requestDto dto.TransactionRequestDto

	err := json.NewDecoder(r.Body).Decode(&requestDto)
	if err != nil {
		log.Error("Json parse error: %w", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validateRequest(w, requestDto); err != nil {
		return
	}

	account, err := c.userRepo.GetAccountByUsername(r.Context(), requestDto.AccountNumber, claims.Username)

	if err != nil {
		log.Error("Error confirm account: %w", err)
		http.Error(w, "Wrong account number", http.StatusBadRequest)
		return
	}

	if account.Status != "active" {
		log.Error("Try to perform trasaction with not active account")
		http.Error(w, "Try to perform trasaction with not active account", http.StatusBadRequest)
		return
	}

	account, err = transaction(r.Context(), *account, requestDto.Amount)

	if err != nil {
		log.Error("transaction error: %w", err)
		http.Error(w, "transaction error", http.StatusBadRequest)
		return
	}

	jsonData, err := json.Marshal(dto.AccountToAccountReponseDto(account))
	if err != nil {
		log.Critical("Encode accounts to json error: %w", err)
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", r.Header.Get("Authorization"))
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
