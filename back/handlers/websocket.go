package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type TypeClient uint8

const (
	//Значения параметров записывающих и читающих подключений
	Reader TypeClient = iota //0
	Writer                   //1
	Model                    //2
)

type Connect struct {
	wsConn *websocket.Conn
	key    string
}

// Создание переменнных карты для сохранения подключений WebSocket
var ReaderClients map[Connect]bool
var WriterClients map[Connect]bool
var ModelClients map[Connect]bool

// Указание размера буфера чтения и записи
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 4,
	WriteBufferSize: 1024 * 4,
}

func WS(c *gin.Context) {
	w := c.Writer
	r := c.Request

	//определение типа
	t, err := strconv.Atoi(c.Query("type"))
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	k := c.Query("key")

	// обновление соединения WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	//Сохранение подключений в карты
	switch {
	case t == int(Reader):
		ReaderClients[Connect{wsConn: ws, key: k}] = true
	case t == int(Writer):
		WriterClients[Connect{wsConn: ws, key: k}] = true
	case t == int(Model):
		ModelClients[Connect{wsConn: ws, key: k}] = true
	default:
		c.Writer.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Неизвестный тип клиента"))
		ws.Close()
		return
	}

	defer delete(ReaderClients, Connect{wsConn: ws, key: k})
	defer delete(WriterClients, Connect{wsConn: ws, key: k})
	defer delete(ModelClients, Connect{wsConn: ws, key: k})
	defer ws.Close()

	for {
		//чтение сообщения
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		switch {
		case WriterClients[Connect{wsConn: ws, key: k}]: //отправка сообщений читающим
			sendToClients(messageType, message, k, ReaderClients)
		}

	}

}

// Отправка сообщений клиентам
func sendToClients(messageType int, message []byte, key string, clients map[Connect]bool) {
	for k := range clients {
		if k.key == key {
			if err := k.wsConn.WriteMessage(messageType, message); err != nil {
				log.Println(err)
				break
			}
		}
	}
}
