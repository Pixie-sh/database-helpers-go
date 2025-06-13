module github.com/pixie-sh/database-helpers-go

go 1.23

toolchain go1.24.1

require (
	github.com/go-gormigrate/gormigrate/v2 v2.1.2
	github.com/google/uuid v1.6.0
	github.com/pixie-sh/errors-go v0.2.1
	github.com/pixie-sh/logger-go v0.1.12
	github.com/pixie-sh/ulid-go v1.1.0
	gorm.io/driver/mysql v1.5.7
	gorm.io/driver/postgres v1.5.7
	gorm.io/gorm v1.30.0
	gorm.io/plugin/dbresolver v1.6.0
)

require (
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.4.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/text v0.20.0 // indirect
)
