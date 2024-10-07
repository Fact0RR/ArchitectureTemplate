package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"server/settings"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Handler struct {
	minio    *minio.Client
	settings settings.Settings
}

func NewHomeHandler(settings settings.Settings) (Handler, error) {
	minioClient, err := minio.New(settings.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(settings.Minio.AccessKeyID, settings.Minio.SecretAccessKey, ""),
		Secure: false, // Установите true, если используете HTTPS
	})
	if err != nil {
		println(settings.Minio.Endpoint)
		println(err.Error())
		return Handler{}, err
	}

	return Handler{minio: minioClient, settings: settings}, nil
}

func (h Handler) Home(c *gin.Context) {
	c.HTML(
		// Установка статуса HTTP на 200
		http.StatusOK,
		// Использование index.html шаблона
		"index.html",
		// Установка title на "Home Page"
		gin.H{
			"title": "hackaton",
		},
	)
}

func (h Handler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		println(http.StatusBadRequest, fmt.Sprintf("Failed to get file: %s", err.Error()))
		c.String(http.StatusBadRequest, fmt.Sprintf("Failed to get file: %s", err.Error()))
		return
	}

	// Открываем файл в памяти
	fileReader, err := file.Open()
	if err != nil {
		println(http.StatusInternalServerError, fmt.Sprintf("Failed to open file: %s", err.Error()))
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to open file: %s", err.Error()))
		return
	}
	defer fileReader.Close()

	// Загрузка файла в MinIO
	_, err = h.minio.PutObject(c, h.settings.Minio.BucketName, file.Filename, fileReader, file.Size, minio.PutObjectOptions{
		ContentType: "video/mp4",
	})
	if err != nil {
		println(http.StatusInternalServerError, fmt.Sprintf("Failed to upload file to MinIO: %s", err.Error()))
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to upload file to MinIO: %s", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file_name": file.Filename,
	})
}

func (h Handler) Stream(c *gin.Context) {
	rawURL := c.Query("url")
	if rawURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter is required"})
		return
	}
	rawKey := c.Query("key")
	if rawURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key parameter is required"})
		return
	}

	// Кодируем URL
	encodedURL := url.QueryEscape(rawURL)
	encodedKEY := url.QueryEscape(rawKey)
	proxyURL := h.settings.Model.Endpoint + "/video_feed?url=" + encodedURL + "&key=" + encodedKEY

	proxyGet(c, proxyURL, "multipart/x-mixed-replace; boundary=frame")
}

func proxyGet(c *gin.Context, proxyURL string, contentType string) {
	// Выполняем GET-запрос
	resp, err := http.Get(proxyURL)
	if err != nil {
		println("Ошибка при отправки GET запроса: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to receive response from source"})
		return
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		println("Статус: ", resp.StatusCode)
		c.JSON(resp.StatusCode, gin.H{"error": "failed to get video feed"})
		return
	}

	// Устанавливаем заголовок Content-Type для ответов с видео-потоком
	c.Header("Content-Type", contentType)

	// Копируем тело ответа от сервиса клиенту
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		println("Ошибка при копировании тела запроса: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stream video"})
		return
	}
}

func (h Handler) Image(c *gin.Context) {
	rawURL := c.Query("url")
	if rawURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter is required"})
		return
	}

	// Кодируем URL
	encodedURL := url.QueryEscape(rawURL)
	proxyURL := h.settings.Model.Endpoint + "/image?url=" + encodedURL

	proxyGet(c, proxyURL, "image/jpeg")
}

func (h Handler) Rtsp(c *gin.Context) {
	rawURL := c.Query("url")
	if rawURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter is required"})
		return
	}
	rawKey := c.Query("key")
	if rawURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key parameter is required"})
		return
	}

	// Кодируем URL
	encodedURL := url.QueryEscape(rawURL)
	encodedKEY := url.QueryEscape(rawKey)
	proxyURL := h.settings.Model.Endpoint + "/rtsp_feed?url=" + encodedURL + "&key=" + encodedKEY
	proxyGet(c, proxyURL, "multipart/x-mixed-replace; boundary=frame")
}
