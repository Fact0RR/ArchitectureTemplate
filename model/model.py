import datetime
import json
import time
from flask import Flask, Response, request
import cv2
from ultralytics import YOLO
import os
from dotenv import load_dotenv
import websocket

# Загружаем переменные окружения из файла .env
load_dotenv()

# Получаем переменные окружения с заданием значений по умолчанию
s3 = os.getenv('ENDPOINT', 'http://minio:9000/test-bucket/')  # Значение по умолчанию
ws_url = os.getenv('WEBSOCKET', 'ws://server:8080/ws?type=1')

app = Flask(__name__)
#model = YOLO('yolov8n.pt')


def gen_frames(urlNameFile,key_websocket):
    # Подключение к WebSocket
    ws = websocket.create_connection(ws_url+'&key='+key_websocket)

    finalUrl = s3 + urlNameFile
    cap = cv2.VideoCapture(finalUrl)
    i = 0
    framePerSecond = 0
    past_time = 0
    frameN = 0
    while True:
        i += 1
        # Преобразуем текущее время в количество секунд с начала эпохи
        current_time_seconds = int(time.time())
        #print(current_time_seconds)
        if current_time_seconds != past_time:
            framePerSecond = frameN
            frameN = 0
            past_time = current_time_seconds
        else:
            frameN = frameN+1

        start_time = time.time()
        ret, frame = cap.read()
        if not ret:
            break
        else:
            ret, buffer = cv2.imencode('.jpg', frame)
            if not ret:
                break
            
            frame_bytes = buffer.tobytes()
            
            yield (b'--frame\r\n'
                   b'Content-Type: image/jpeg\r\n\r\n' + frame_bytes + b'\r\n')
            end_time = time.time()

            # Вычисляем время выполнения
            execution_time = end_time - start_time
            
            
            current_time = datetime.datetime.now().time()
            t = f"{current_time.hour}:{current_time.minute}:{current_time.second}"
            mess = json.dumps({"Номер кадра":i,"Время": t,"Количество обработанных кадров в секунду":int(1/execution_time),"Итоговое количество кадров в секунду":int(framePerSecond)})

            # Отправляем mess в WebSocket
            ws.send(mess)
    
    # Закрываем соединение WebSocket
    ws.close()
    cap.release()

@app.route('/video_feed')
def video_feed():
    url_name_file = request.args.get('url')
    key_websocket = request.args.get('key')
    return Response(gen_frames(url_name_file,key_websocket), mimetype='multipart/x-mixed-replace; boundary=frame')

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)