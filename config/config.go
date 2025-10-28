package config

type Config struct {
	Database DatabaseConfig
	NATS     NATSConfig
	HTTP     HTTPConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type NATSConfig struct {
	ClusterID string
	ClientID  string
	URL       string
	Subject   string
}

type HTTPConfig struct {
	Port string
}

func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5433",
			User:     "wbuser",
			Password: "wbpassword",
			DBName:   "wb_orders",
			SSLMode:  "disable",
		},
		NATS: NATSConfig{
			ClusterID: "test-cluster",
			ClientID:  "wb-orders-service",
			URL:       "nats://localhost:4222",
			Subject:   "orders",
		},
		HTTP: HTTPConfig{
			Port: "8080",
		},
	}
}
