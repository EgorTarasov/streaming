FROM python:3.11.8

WORKDIR /code
RUN apt-get update && apt-get install ffmpeg libsm6 libxext6  -y
COPY requirements.txt requirements.txt
RUN pip3 install -r requirements.txt
COPY . /code