import cv2
import sys
import signal
import redis
import numpy as np
from src.infrastructure.thermal_engine import OpenCVThermalEngine
from src.infrastructure.redis.redis_storage import RedisStorage
from src.infrastructure.video.streamer import FFmpegStreamer
from src.usecase.process_stream import StreamProcessor


CAMERA_INDEX = 2
WIDTH, HEIGHT = 256, 192


CAL_LOW_PIXEL  = 50
CAL_LOW_TEMP   = 20.0

CAL_HIGH_PIXEL = 200
CAL_HIGH_TEMP  = 37.0


def open_camera(index: int, width: int, height: int) -> cv2.VideoCapture:
    cap = cv2.VideoCapture(index, cv2.CAP_V4L2)

    cap.set(cv2.CAP_PROP_FOURCC, cv2.VideoWriter_fourcc(*'Y16 '))
    cap.set(cv2.CAP_PROP_CONVERT_RGB, 0)
    cap.set(cv2.CAP_PROP_FRAME_WIDTH, width)
    cap.set(cv2.CAP_PROP_FRAME_HEIGHT, height)

    if not cap.isOpened():
        raise RuntimeError(f"Не удалось открыть камеру /dev/video{index}")

    fourcc_val = int(cap.get(cv2.CAP_PROP_FOURCC))
    fourcc_str = "".join([chr((fourcc_val >> (8 * i)) & 0xFF) for i in range(4)])
    actual_w = int(cap.get(cv2.CAP_PROP_FRAME_WIDTH))
    actual_h = int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT))

    print(f"[Camera] FOURCC={fourcc_str!r}  {actual_w}×{actual_h}")

    if fourcc_str.strip() not in ('Y16',):
        print("[Camera] Y16 недоступен — работаем через YUYV (псевдоцвет).")
        print("         Точность температур зависит от калибровки.")
        print("         Проверь форматы: v4l2-ctl --list-formats-ext -d /dev/video2")

    return cap



def read_frame(cap: cv2.VideoCapture) -> np.ndarray | None:
    """
    Читает кадр и нормализует формат:
      - YUYV (2 канала) → BGR
      - Y16 (uint16)   → остаётся uint16
      - BGR / Gray     → без изменений
    """
    ret, frame = cap.read()
    if not ret or frame is None:
        return None

    if frame.ndim == 3 and frame.shape[2] == 2:
        frame = cv2.cvtColor(frame, cv2.COLOR_YUV2BGR_YUYV)

    return frame


def make_gray_for_stream(frame: np.ndarray) -> np.ndarray:
    """
    Строит 8-bit grayscale для передачи в FFmpegStreamer.
    FFmpeg настроен на -pix_fmt gray → нужен ровно 1 канал uint8.
    """
    if frame.dtype == np.uint16:
        out = cv2.normalize(frame, None, 0, 255, cv2.NORM_MINMAX)
        return out.astype(np.uint8)

    if frame.ndim == 3 and frame.shape[2] == 3:
        return cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)

    if frame.ndim == 3 and frame.shape[2] == 2:
        return frame[:, :, 0]  # Y из YUYV

    return frame if frame.dtype == np.uint8 else frame.astype(np.uint8)




def main():
    cap = open_camera(CAMERA_INDEX, WIDTH, HEIGHT)

    engine = OpenCVThermalEngine(
        grid_size=(8, 6),
        cal_low=(CAL_LOW_PIXEL, CAL_LOW_TEMP),
        cal_high=(CAL_HIGH_PIXEL, CAL_HIGH_TEMP),
    )

    _client = redis.Redis(host='127.0.0.1', port=6379, decode_responses=True)
    storage = RedisStorage(_client)

    streamer = FFmpegStreamer(
        rtmp_url="rtmp://127.0.0.1:1935/cam1_analyzed",
        width=WIDTH,
        height=HEIGHT,
    )

    processor = StreamProcessor(engine, storage, streamer)

    def signal_handler(sig, frame):
        print("\nЗавершение работы...")
        cap.release()
        streamer.close()
        sys.exit(0)

    signal.signal(signal.SIGINT, signal_handler)
    print("Система запущена. Нажми Ctrl+C для остановки.")

    consecutive_errors = 0
    MAX_ERRORS = 30

    while True:
        frame = read_frame(cap)

        if frame is None:
            consecutive_errors += 1
            if consecutive_errors >= MAX_ERRORS:
                print(f"[ERROR] Камера не отвечает {MAX_ERRORS} кадров подряд — выход.")
                break
            continue

        consecutive_errors = 0

        processor.execute(frame)



if __name__ == "__main__":
    main()