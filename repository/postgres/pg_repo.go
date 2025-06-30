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

// PRIVATE SECTION
