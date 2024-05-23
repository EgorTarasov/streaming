import logging
import cv2
import numpy as np

from ultralytics import YOLO
import json
import base64
from kafka import KafkaConsumer, KafkaProducer
from typing import TypedDict
from ultralytics.utils.plotting import Annotator
import time
import os

PREDICTION_TOPIC = os.getenv("PREDICTION_TOPIC", "prediction")
IMAGES_TOPIC = os.getenv("IMAGES_TOPIC", "frames-splitted")
COMMAND_TOPIC = os.getenv("COMMAND_TOPIC", default="prediction")
CMD_RESPONSE_TOPIC = os.getenv("CMD_RESPONSE_TOPIC", default="responses")
KAFKA_BROKERS = os.getenv("KAFKA_BROKERS", "localhost:9091")
MODEL_PATH = os.getenv("YOLO_PATH", default="./models/yolov8n.pt")


logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)
formatter = logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")
ch = logging.StreamHandler()
ch.setFormatter(formatter)
logger.addHandler(ch)


logger.info(PREDICTION_TOPIC)
logger.info(IMAGES_TOPIC)
logger.info(KAFKA_BROKERS)


def custom_deserializer(m):
    m = m.decode("utf-8")
    m_json = json.loads(m)
    m_json["frame"] = base64.b64decode(m_json["frame"])  # Decode base64 string to bytes
    return m_json


def custom_serializer(m):
    m["frame"] = base64.b64encode(m["frame"]).decode(
        "utf-8"
    )  # Encode bytes to base64 string
    return json.dumps(m).encode("utf-8")


image_consumer = KafkaConsumer(
    IMAGES_TOPIC,
    bootstrap_servers=KAFKA_BROKERS,
    value_deserializer=custom_deserializer,
)


class ResultMessage(TypedDict):
    streamId: int
    frameId: str
    frame: str  # base64 encoded image
    results: dict  # YOLO results


producer = KafkaProducer(
    bootstrap_servers=KAFKA_BROKERS,
    value_serializer=lambda v: json.dumps(v).encode("utf-8"),
)

# Load a pretrained YOLO model
logger.info(f"loading yolov from {MODEL_PATH}")
model = YOLO(MODEL_PATH, verbose=False)
logger.info("loaded ml")

for msg in image_consumer:
    start = time.time()
    logger.info(f"Received message from {IMAGES_TOPIC}")

    nparr = np.fromstring(msg.value["frame"], np.uint8)
    img_np = cv2.imdecode(nparr, cv2.IMREAD_COLOR)

    # Pass the numpy image to the YOLO model
    results = model(img_np)
    for r in results:
        annotator = Annotator(img_np)

        boxes = r.boxes
        for box in boxes:
            b = box.xyxy[0]
            c = box.cls
            annotator.box_label(b, model.names[int(c)])

    img_np = annotator.result()
    resultMsg = ResultMessage(
        streamId=msg.value["streamId"],
        frameId=msg.value["frameId"],
        frame=base64.b64encode(cv2.imencode(".jpg", img_np)[1]).decode("utf-8"),
        results={},  # box.tojson(normalize=True) for box in results
    )
    producer.send(PREDICTION_TOPIC, resultMsg)
    logger.info(f"Sent message to {PREDICTION_TOPIC} in {time.time()-start} seconds")
