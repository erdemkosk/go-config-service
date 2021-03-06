package plugin

import (
	"os"

	models "github.com/erdemkosk/go-config-service/api/models"
	_ "github.com/joho/godotenv/autoload"
)

var (
	Config models.ServerConfiguration
)

func GetEnvConfig(key string) string {
	return os.Getenv(key)
}
