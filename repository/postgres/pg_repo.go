package postgres

import (
	"context"
	"database/sql"
	"time"
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
}

func PgConfigFromConfig(cfg utils.Config) PgConfig {
	return PgConfig{
		Host:       cfg.DbHost,
		Port:       cfg.DbPort,
		User:       cfg.DbUsername,
		Password:   cfg.DbPassword,
		Name:       cfg.DbName,
		CtxSecTout: cfg.DbCtxTimeoutSec,
	}
}

func (cfg *PgConfig) String() string {
	return "host=" + cfg.Host +
		" port=" + cfg.Port +
		" user=" + cfg.User +
		" password=" + cfg.Password +
		" dbname=" + cfg.Name
}

func New(ctx context.Context, cfg PgConfig) *PostgresRepository {
	log := utils.GlobalLogger()
	log.Info("Try to connect to the Postgres DB...")

	connCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.CtxSecTout)*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", cfg.String())
	if err != nil {
		log.Critical("Cann't connect to to db: %w", err)
		return nil
	}

	if err := db.PingContext(connCtx); err != nil {
		log.Critical("Cann't ping db: %w", err)
		db.Close()
		return nil
	}

	log.Info("Postgres Ping OK!")
	/*if err := initSchema(ctx, db); err != nil {
		log.Critical("Cann't init db schema: %w", err)
		db.Close()
		return nil
	}*/

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

// PRIVATE SECTION
