//генерация ключа для websocket
function getCurrentDateTimeFormatted() {
    const now = new Date();

    const year = now.getFullYear().toString();
    const month = String(now.getMonth() + 1).padStart(2, '0'); // Месяцы начинаются с 0
    const day = String(now.getDate()).padStart(2, '0');
    const hours = String(now.getHours()).padStart(2, '0');
    const minutes = String(now.getMinutes()).padStart(2, '0');
    const seconds = String(now.getSeconds()).padStart(2, '0');
    const milliseconds = String(now.getMilliseconds()).padStart(3, '0'); // Добавлено для милисекунд

    return `${year}${month}${day}${hours}${minutes}${seconds}${milliseconds}`;
}

const GlobalKey = getCurrentDateTimeFormatted()

//подключение, как слушащий канал
let socket = new WebSocket("ws://localhost:8080/ws?type=0&key="+GlobalKey);

//уведомляем, что websocket подключен
socket.onopen = function(event) {
    console.log("connected");
};

//карта для названий колонок в результирующей таблице
let currentMap = new Map();

//измение карты, если присылается json другой структуры
function clearMapAndSetNewValues(keysArr,obj){
  currentMap.clear();
  reBuildTable(keysArr,obj);
  setNewValues(keysArr,obj);
}

//пересборка таблицы
function reBuildTable(keysArr,obj){
    const table = document.querySelector('table');
    while (table.rows.length) {
        table.deleteRow(0);
    }

    var head = document.createElement("tr")
    var body = document.createElement("tr")

    for (let i = 0;i< keysArr.length;i++){
        var th = document.createElement("th");
        var td = document.createElement("td");
        
        td.id = keysArr[i]

        th.textContent = keysArr[i]
        td.textContent = obj[keysArr[i]]

        head.appendChild(th);
        body.appendChild(td);
    }

    table.appendChild(head)
    table.appendChild(body)

}


function setNewValues(keysArr,obj){
  for (var i = 0;i<keysArr.length;i++)
  {
      currentMap.set(keysArr[i],obj[keysArr[i]])
  }
}

//установка таблицы
function setTable(obj){

    keysArr = Object.keys(obj);

    if (currentMap.size!==keysArr.length)
    {
        clearMapAndSetNewValues(keysArr,obj)
    }

    for (var i = 0;i<keysArr.length;i++)
    {
        if (currentMap.get(keysArr[i]) == null)
        {
            clearMapAndSetNewValues(keysArr,obj) 
        }
    }
    
    updateData(keysArr,obj)
}

function updateData(keysArr,obj){
    for (var i = 0;i<keysArr.length;i++)
    {
        document.getElementById(keysArr[i]).textContent = obj[keysArr[i]]
    }
}

//получение данных с сервера
socket.onmessage = function(event) {
    obj = JSON.parse(event.data);
    b64 = obj["b64"]

    if (b64!=undefined){
        //отрисовка изображения на экран
        document.getElementById("content").src = b64
        //удаление данных, чтобы код изображения не попал в таблицу
        delete(obj["b64"]);
    }

    setTable(obj);
};

document.getElementById('videoUploadForm').addEventListener('submit', function(event) {
    event.preventDefault(); // Отменяем стандартное поведение формы

    const videoInput = document.getElementById('video');
    const message = document.getElementById('message');

    if (videoInput.files.length === 0) {
        message.textContent = 'Пожалуйста, выберите видео для загрузки.';
        message.style.color = 'red';
        return;
    }

    // Здесь вы можете добавить код для отправки видео на сервер (например, с использованием Fetch API)

    // Для примера, просто показываем сообщение о успешной загрузке
    message.textContent = 'Видео успешно загружено!';
    message.style.color = 'green';

});

document.getElementById('uploadButton').addEventListener('click', uploadFile);

async function uploadFile() {
    const fileInput = document.getElementById('video');
    const file = fileInput.files[0];
    const outputImage = document.getElementById('outputImage');

    if (!file) {
        alert('Пожалуйста, выберите файл для загрузки.');
        return;
    }

    const formData = new FormData();
    formData.append('file', file);

    fetch('/upload', {
        method: 'POST',
        body: formData
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Ошибка при загрузке файла.');
        }
        return response.json();
    })
    .then(data => {
        // Проверяем тип файла
        const type = file.type;
        if (type.startsWith('video/')) {
            outputImage.src = '/stream?url='+file.name+'&key='+GlobalKey; // Замените на адрес вашего изображения
            outputImage.style.display = 'block'; // Показываем изображение
        } else if (type.startsWith('image/')) {
            outputImage.src = '/image?url='+file.name;
            outputImage.style.display = 'block'; // Показываем изображение
        } else {
            alert('Выбранный файл не является видео или изображением.');
            return
        }

        outputImage.onload = function() { // Ждём загрузку изображения
            outputImage.scrollIntoView({ behavior: 'smooth', block: 'start' });
        };

    })
    .catch(error => {
        console.error('Ошибка:', error);
        alert('Произошла ошибка при загрузке.');
    });
}

document.getElementById('rtspButton').addEventListener('click', async () => {
    const urlRTSP = document.getElementById('urlRTSP').value;
    const outputImage = document.getElementById('outputImage');

    // Проверка на пустое текстовое поле
    if (!urlRTSP) {
        alert('Пожалуйста, введите RTSP URL.');
        return; // Прерываем выполнение функции, если поле пустое
    }

    // Формируем URL с параметрами
    const endpoint = `/rtsp?key=${encodeURIComponent(GlobalKey)}`+`&url=${encodeURIComponent(urlRTSP)}`;
    console.log(endpoint)
    outputImage.src = endpoint
    outputImage.style.display = 'block'; // Показываем изображение
    outputImage.onload = function() { // Ждём загрузку изображения
        outputImage.scrollIntoView({ behavior: 'smooth', block: 'start' });
    };
});