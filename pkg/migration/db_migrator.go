package migrations

// DBMigrator is simple database migrator interface
type DBMigrator interface {
	Initialize() error
	Up() error
	Down() error
}
