package main

import (
	"log"
	"server/handlers"
	"server/settings"

	"github.com/gin-gonic/gin"
)

func main() {

	handlers.ReaderClients = make(map[handlers.Connect]bool)
	handlers.WriterClients = make(map[handlers.Connect]bool)
	handlers.ModelClients = make(map[handlers.Connect]bool)

	settings, err := settings.NewSettings()
	if err != nil {
		if err.Error() == "Ошибка при загрузке .env файла open .env: The system cannot find the file specified." {
			log.Println(err.Error())
		} else {
			log.Println(err.Error())
		}
	}
	//Подключаем все файлы связанные с фронтендом
	r := gin.Default()
	r.LoadHTMLGlob("./front/*")
	r.StaticFile("/script.js", "./front/script.js")
	r.StaticFile("/style.css", "./front/style.css")
	r.StaticFile("/favicon.ico", "./front/icon.ico")
	homeHandler, err := handlers.NewHomeHandler(settings)
	if err != nil {
		log.Fatal(err.Error())
	}
	//Добавляем маршруты для домашней страницы и для подключения
	r.GET("/", homeHandler.Home)
	r.POST("/upload", homeHandler.Upload)
	r.GET("/stream", homeHandler.Stream)
	r.GET("/rtsp", homeHandler.Rtsp)
	r.GET("/image", homeHandler.Image)
	r.GET("/ws", handlers.WS)
	//Запускаем сервер на потру 8080
	r.Run(":8080")
}
