package settings

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Settings struct {
	Minio MinioSeettings
	Model ModelSettings
}

func NewSettings() (Settings, error) {
	err := godotenv.Load()
	if err != nil {
		return Settings{
			Minio: MinioSeettings{
				BucketName:      getEnv("BUCKET_NAME", "test-bucket"),
				Endpoint:        getEnv("ENDPOINT", "minio:9000"),
				AccessKeyID:     getEnv("ACCESS_KEY_ID", "minio-user"),
				SecretAccessKey: getEnv("SECRET_ACCESS_KEY", "minio-password"),
			},
			Model: ModelSettings{
				Endpoint: getEnv("ENDPOINT_MODEL", "http://model:5000"),
			},
		}, errors.New("Ошибка при загрузке .env файла " + err.Error())
	}

	// Инициализация конфигурации с дефолтными значениями
	sett := Settings{
		Minio: MinioSeettings{
			BucketName:      getEnv("BUCKET_NAME", "test-bucket"),
			Endpoint:        getEnv("ENDPOINT", "127.0.0.1:9000"),
			AccessKeyID:     getEnv("ACCESS_KEY_ID", "minio-user"),
			SecretAccessKey: getEnv("SECRET_ACCESS_KEY", "minio-password"),
		},
		Model: ModelSettings{
			Endpoint: getEnv("ENDPOINT_MODEL", "http://model:5000"),
		},
	}
	return sett, nil
}

// getEnv получает значение переменной окружения или возвращает дефолтное значение
func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

type MinioSeettings struct {
	BucketName      string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
}

type ModelSettings struct {
	Endpoint string
}
