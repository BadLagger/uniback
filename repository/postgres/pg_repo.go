package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"
	"uniback/dto"
	"uniback/models"
	"uniback/utils"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

// Минимальный набор
type PgConfig struct {
	Host       string
	Port       string
	User       string
	Password   string
	Name       string
	CtxSecTout int
	SslMode    string
}

func PgConfigFromConfig(cfg utils.Config) PgConfig {
	sslModeStr := "disable"
	if cfg.DbSslMode {
		sslModeStr = "enable"
	}
	return PgConfig{
		Host:       cfg.DbHost,
		Port:       cfg.DbPort,
		User:       cfg.DbUsername,
		Password:   cfg.DbPassword,
		Name:       cfg.DbName,
		CtxSecTout: cfg.DbCtxTimeoutSec,
		SslMode:    sslModeStr,
	}
}

func (cfg *PgConfig) String() string {
	return "host=" + cfg.Host +
		" port=" + cfg.Port +
		" user=" + cfg.User +
		" password=" + cfg.Password +
		" dbname=" + cfg.Name +
		" sslmode=" + cfg.SslMode
}

func New(ctx context.Context, cfg PgConfig) *PostgresRepository {
	log := utils.GlobalLogger()
	log.Info("Try to connect to the Postgres DB...")

	connCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.CtxSecTout)*time.Second)
	defer cancel()
	log.Debug("SSL open string: %s", cfg.String())
	db, err := sql.Open("postgres", cfg.String())
	if err != nil {
		log.Critical("Cann't connect to to db: %w", err)
		return nil
	}

	if err := db.PingContext(connCtx); err != nil {
		log.Critical("Cann't ping db")
		if opErr, ok := err.(*net.OpError); ok {
			log.Trace("Operation: %s", opErr.Op)
			log.Trace("Net: %s", opErr.Net)
			log.Trace("Address: %v", opErr.Addr)
			log.Trace("Error: %s", opErr.Err)

			if opErr.Err != nil {
				log.Trace("Error details: %+v", opErr.Err)
			}
		} else {
			log.Trace("DbError: %w", err)
		}
		db.Close()
		return nil
	}

	log.Info("Postgres Ping OK!")
	if err := initSchema(ctx, db); err != nil {
		log.Critical("Cann't init db schema: %w", err)
		db.Close()
		return nil
	}

	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Close() error {
	utils.GlobalLogger().Info("Close DB!")
	return r.db.Close()
}

func (r *PostgresRepository) Ping(ctx context.Context) error {
	utils.GlobalLogger().Info("Ping DB!")
	return r.db.PingContext(ctx)
}

func (r *PostgresRepository) IsUserExistsByUsernameEmailPhone(ctx context.Context, userDto dto.UserCreateRequest) (username, email, phone bool, err error) {
	query := `
		SELECT
			EXISTS(SELECT 1 FROM users WHERE username = $1),
			EXISTS(SELECT 1 FROM users WHERE email    = $2),
			EXISTS(SELECT 1 FROM users WHERE phone    = $3)`

	utils.GlobalLogger().Debug("Try to check Username %s with Email %s and Phone %s", userDto.Username, userDto.Email, userDto.Phone)
	err = r.db.QueryRowContext(ctx, query, userDto.Username, userDto.Email, userDto.Phone).Scan(
		&username,
		&email,
		&phone,
	)
	if err != nil {
		return false, false, false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return username, email, phone, nil
}

func (r *PostgresRepository) CreateUser(ctx context.Context, userDto dto.UserCreateRequest) error {
	query := `
		INSERT INTO 
			users (username, password, email, phone)
		VALUES ($1, $2, $3, $4)
	`

	utils.GlobalLogger().Debug("Try to create %s with Email %s and Phone %s", userDto.Username, userDto.Email, userDto.Phone)

	_, err := r.db.ExecContext(ctx, query, userDto.Username, userDto.Password, userDto.Email, userDto.Phone)

	return err
}

func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT
			id, username, password, email, phone
		FROM
			users
		WHERE username = $1
	`

	var user models.User

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Name,
		&user.Password,
		&user.Email,
		&user.Phone,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}

func (r *PostgresRepository) IsUserExists(ctx context.Context, username string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM users WHERE username = $1
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, username).Scan(&count)
	if err != nil {
		return false, err
	}

	if count == 0 {
		return false, nil
	}

	if count > 1 {
		return false, fmt.Errorf("Dublicated names in DB: %s (count = %d)", username, count)
	}

	return true, nil
}

func (r *PostgresRepository) GetUserId(ctx context.Context, username string) (int, error) {
	query := `
		SELECT
			id
		FROM
			users
		WHERE username = $1
	`

	var userId int

	err := r.db.QueryRowContext(ctx, query, username).Scan(&userId)
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func (r *PostgresRepository) GetAccountsByUsername(ctx context.Context, username string) (*dto.AccountsResponseDto, error) {
	const query = `
        SELECT a.account_number, a.account_type, 
               a.balance, a.opening_date, a.status
        FROM accounts a
        JOIN users u ON a.user_id = u.id
        WHERE u.username = $1
        ORDER BY a.opening_date DESC
    `

	rows, err := r.db.Query(query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to query user accounts: %w", err)
	}
	defer rows.Close()

	var accounts []dto.AccountResponseDto
	for rows.Next() {
		var acc dto.AccountResponseDto
		err := rows.Scan(
			&acc.AccountNumber,
			&acc.AccountType,
			&acc.Balance,
			&acc.OpeningDate,
			&acc.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, acc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	response := dto.AccountsResponseDto{
		AccountsNum: len(accounts),
		Acounts:     accounts,
	}

	return &response, nil
}

func (r *PostgresRepository) IsAccountExits(ctx context.Context, accountNumber string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM accounts WHERE account_number = $1
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, accountNumber).Scan(&count)
	if err != nil {
		return false, err
	}

	if count == 0 {
		return false, nil
	}

	if count > 1 {
		return false, fmt.Errorf("Dublicated names in DB: %s (count = %d)", accountNumber, count)
	}

	return true, nil
}

func (r *PostgresRepository) GetAccountByNumber(ctx context.Context, number string) (*models.Account, error) {
	query := `
		SELECT
    		id, user_id, account_number, account_type, balance, opening_date, status
		FROM
    		accounts
		WHERE
			account_number = $1
	`

	var Account models.Account
	err := r.db.QueryRowContext(ctx, query, number).Scan(
		&Account.Id,
		&Account.UserId,
		&Account.AccountNumber,
		&Account.AccountType,
		&Account.Balance,
		&Account.OpeningDate,
		&Account.Status)
	if err != nil {
		return nil, err
	}

	return &Account, nil
}

func (r *PostgresRepository) CreateAccount(ctx context.Context, acc models.Account) (*dto.AccountResponseDto, error) {
	query := `
		INSERT INTO 
			accounts (user_id, account_number, account_type, status)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query, acc.UserId, acc.AccountNumber, acc.AccountType, acc.Status)

	if err != nil {
		return nil, err
	}

	account, err := r.GetAccountByNumber(ctx, acc.AccountNumber)
	if err != nil {
		return nil, err
	}

	return dto.AccountToAccountReponseDto(account), nil
}

func (r *PostgresRepository) GetAccountByUsername(ctx context.Context, account string, username string) (*models.Account, error) {
	query := `
		SELECT
			a.id, a.user_id, a.account_number, a.account_type, a.balance, a.opening_date, a.status
		FROM
			accounts a
			JOIN users u ON a.user_id = u.id
		WHERE
			u.username = $1 AND a.account_number = $2
	`

	var resultAccount models.Account

	err := r.db.QueryRowContext(ctx, query, username, account).Scan(
		&resultAccount.Id,
		&resultAccount.UserId,
		&resultAccount.AccountNumber,
		&resultAccount.AccountType,
		&resultAccount.Balance,
		&resultAccount.OpeningDate,
		&resultAccount.Status)

	if err != nil {
		return nil, err
	}

	return &resultAccount, nil
}

func (r *PostgresRepository) UpdateAccountTransaction(ctx context.Context, acc models.Account, amount float64, fee float64, trsType string) (*models.Account, error) {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		"UPDATE accounts SET balance = $1 WHERE id = $2",
		acc.Balance, acc.Id,
	)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec(
		"INSERT INTO transactions (account_id, type, amount, fee) VALUES($1, $2 , $3, $4)",
		acc.Id,
		trsType,
		amount,
		fee,
	)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	result, err := r.GetAccountByNumber(ctx, acc.AccountNumber)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *PostgresRepository) TransferAccountsTransaction(ctx context.Context, src models.Account, dest models.Account, amount float64, fee float64) (*models.Account, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		"UPDATE accounts SET balance = $1 WHERE id = $2",
		src.Balance, src.Id,
	)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec(
		"UPDATE accounts SET balance = $1 WHERE id = $2",
		dest.Balance, dest.Id,
	)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var transactionId int

	err = tx.QueryRow(
		"INSERT INTO transactions (account_id, type, amount, fee) VALUES($1, 'transfer', $2, $3) RETURNING id",
		src.Id,
		amount,
		fee,
	).Scan(&transactionId)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec(
		"INSERT INTO transaction_trasfers (trans_id, dest_account_id) VALUES($1, $2)",
		transactionId, dest.Id,
	)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	result, err := r.GetAccountByNumber(ctx, src.AccountNumber)

	if err != nil {
		return nil, err
	}

	return result, nil

}

func (r *PostgresRepository) IsCardExists(ctx context.Context, number []byte) (bool, error) {
	query := `
		SELECT COUNT(*) FROM cards WHERE number = $1
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, number).Scan(&count)
	if err != nil {
		return false, err
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

func (r *PostgresRepository) CreateNewCard(ctx context.Context, card models.Card) error {
	query := `
		INSERT INTO 
			cards (account_id, number, expiry, cvv, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query, card.AccountId, card.Number, card.Expiry, card.Cvv, card.CreatedAt)

	return err
}

// PRIVATE SECTION
