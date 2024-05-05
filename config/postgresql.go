package config

var PostgresqlConfig = ContainerConfig{
	Name:  "snapshot-testing-postgresql",
	Image: "timescale/timescaledb:2.8.0-pg14",
	// Environment: map[string]string{
	// 	"POSTGRES_USER":     "vega",
	// 	"POSTGRES_DB":       "vega",
	// 	"POSTGRES_PASSWORD": "vega",
	// },
	Command: []string{
		"postgres",
		"-c", "max_connections=50",
		"-c", "log_destination=stderr",
		"-c", "work_mem=5MB",
		"-c", "huge_pages=off",
		"-c", "shared_memory_type=sysv",
		"-c", "dynamic_shared_memory_type=sysv",
		"-c", "shared_buffers=2GB",
		"-c", "temp_buffers=5MB",
	},
	// Ports: map[uint16]uint16{
	// 	5432: 5432,
	// },
}

// TODO: For now we are hardcoding it
var DefaultCredentials = PostgreSQLCreds{
	Host:   "localhost",
	Port:   5432,
	User:   "vega",
	Pass:   "vega",
	DbName: "vega",
}
