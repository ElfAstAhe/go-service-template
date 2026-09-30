package trace

import (
	"github.com/ElfAstAhe/go-service-template/pkg/config"
	"github.com/ElfAstAhe/go-service-template/pkg/db/postgres"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/XSAM/otelsql"
	"github.com/xo/dburl"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// New конструктор экземпляра db.DB совместно с трассировкой
func New(conf *config.DBConfig) (*postgres.DB, error) {
	u, err := dburl.Parse(conf.DSN)
	if err != nil {
		return nil, errs.NewDalError("postgres/trace/New", "parse DSN", err)
	}
	// Вместо sql.Open("postgres", ...) делаем:
	pg, err := otelsql.Open("postgres", conf.DSN,
		otelsql.WithAttributes(
			semconv.DBSystemNamePostgreSQL,
			semconv.DBSystemNameKey.String(u.Path),
		),
		// Включаем трейсинг всех запросов к БД
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			Ping: true, // Трейсить даже проверки связи (healthchecks)
		}),
	)
	if err != nil {
		return nil, errs.NewDalError("postgres/trace/New", "failed to open otel pgx db", err)
	}

	appDB, err := postgres.Setup(pg, conf)
	if err != nil {
		return nil, errs.NewDalError("postgres/trace/New", "failed setup db connection", err)
	}

	return appDB, nil
}
