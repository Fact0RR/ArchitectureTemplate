FROM pytorch/pytorch:2.1.0-cuda12.1-cudnn8-runtime

# Add requirements
ADD ./requirements.txt ./requirements.txt

# Install requirements
RUN pip install -r requirements.txt

RUN apt-get update && apt-get install ffmpeg libsm6 libxext6  -y
# Add files
ADD . .

EXPOSE 5000

ENTRYPOINT ["python3", "model.py"]