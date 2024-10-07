import datetime
from io import BytesIO
import json
import time
from flask import Flask, Response, jsonify, request, send_file
import cv2
import numpy as np
import requests
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


def gen_video(urlNameFile,key_websocket,rtsp):
    if rtsp:
        url = urlNameFile
    else:
        url = s3 + urlNameFile
    # Подключение к WebSocket
    ws = websocket.create_connection(ws_url+'&key='+key_websocket)
    cap = cv2.VideoCapture(url)
    i = 0
    framePerSecond = 0
    past_time = 0
    frameN = 0
    while True:
        i += 1
        # Преобразуем текущее время в количество секунд с начала эпохи
        current_time_seconds = int(time.time())
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
            ws.send(mess)
            # Отправляем mess в WebSocket
            # Закрываем соединение WebSocket
    ws.close()
    cap.release()

@app.route('/image', methods=['GET'])
def get_image():
    
    url_name_file = request.args.get('url')

    print(s3+url_name_file)
    # Получаем изображение по URL
    response = requests.get(s3+url_name_file)

    # Проверяем статус ответа
    if response.status_code != 200:
        return jsonify({'error': 'Не удалось получить изображение по указанному URL'}), 400

    # Преобразуем содержимое ответа в массив NumPy
    image_array = np.asarray(bytearray(response.content), dtype=np.uint8)
    
    # Декодируем массив в изображение OpenCV
    image = cv2.imdecode(image_array, cv2.IMREAD_COLOR)

    if image is None:
        return jsonify({'error': 'Не удалось декодировать изображение'}), 400
        
    # Кодируем изображение в формат JPEG и сохраняем в буфер
    _, encoded_image = cv2.imencode('.jpg', image)
    byte_img = BytesIO(encoded_image.tobytes())  # Преобразуем в поток байтов

    return send_file(byte_img, mimetype='image/jpeg')  # Возвращаем изображение в байтах

@app.route('/rtsp_feed', methods=['GET'])
def rtsp_feed(): 
    url_name_file = request.args.get('url')
    key_websocket = request.args.get('key')
    return Response(gen_video(url_name_file,key_websocket,rtsp=True), mimetype='multipart/x-mixed-replace; boundary=frame')

@app.route('/video_feed', methods=['GET'])
def video_feed():
    url_name_file = request.args.get('url')
    key_websocket = request.args.get('key')
    return Response(gen_video(url_name_file,key_websocket,rtsp=False), mimetype='multipart/x-mixed-replace; boundary=frame')

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)