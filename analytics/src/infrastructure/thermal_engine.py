import numpy as np
import cv2
from src.domain.interfaces import ThermalEngine
from src.domain.models import ThermalStats


class OpenCVThermalEngine(ThermalEngine):
    DEBUG_PIXELS = False

    def __init__(
        self,
        grid_size: tuple[int, int] = (8, 6),
        cal_low: tuple[int, float] = (50, 20.0),
        cal_high: tuple[int, float] = (200, 37.0),
        clip_min: float = -10.0,
        clip_max: float = 120.0,
    ):
        self.grid_w, self.grid_h = grid_size
        self.clip_min = clip_min
        self.clip_max = clip_max

        p1, t1 = cal_low
        p2, t2 = cal_high
        if p2 == p1:
            raise ValueError("cal_low и cal_high должны иметь разные значения пикселей")
        self._a = (t2 - t1) / (p2 - p1)
        self._b = t1 - self._a * p1

        print(
            f"[ThermalEngine] Калибровка: T = {self._a:.4f} × pixel + {self._b:.2f}  "
            f"| {p1}px={t1}°C  {p2}px={t2}°C"
        )


    def _to_gray(self, frame: np.ndarray) -> np.ndarray:
        """Приводит любой формат кадра к 8-bit grayscale."""
        if frame.dtype == np.uint16:
            gray = cv2.normalize(frame, None, 0, 255, cv2.NORM_MINMAX)
            return gray.astype(np.uint8)

        if frame.ndim == 2:
            return frame if frame.dtype == np.uint8 else frame.astype(np.uint8)

        if frame.ndim == 3:
            ch = frame.shape[2]
            if ch == 3:
                return cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
            if ch == 2:
                return frame[:, :, 0]
            if ch == 4:
                return cv2.cvtColor(frame, cv2.COLOR_BGRA2GRAY)

        raise ValueError(f"Неподдерживаемый формат кадра: shape={frame.shape} dtype={frame.dtype}")

    def _pixel_to_celsius(self, gray: np.ndarray) -> np.ndarray:
        """Переводит 8-bit яркость → градусы Цельсия через калибровочную прямую."""
        celsius = self._a * gray.astype(np.float32) + self._b
        return np.clip(celsius, self.clip_min, self.clip_max)

    def analyze(self, frame: np.ndarray) -> ThermalStats:
        gray = self._to_gray(frame)

        if self.DEBUG_PIXELS:
            print(
                f"[DEBUG] pixel: min={gray.min()}  max={gray.max()}  "
                f"mean={gray.mean():.1f}  center={gray[gray.shape[0]//2, gray.shape[1]//2]}"
            )

        celsius = self._pixel_to_celsius(gray)

        orig_h, orig_w = gray.shape[:2]

        grid = cv2.resize(celsius, (self.grid_w, self.grid_h), interpolation=cv2.INTER_AREA)

        min_val, max_val, min_loc, max_loc = cv2.minMaxLoc(grid)

        cell_w = orig_w // self.grid_w
        cell_h = orig_h // self.grid_h
        center_x = max_loc[0] * cell_w + cell_w // 2
        center_y = max_loc[1] * cell_h + cell_h // 2
        box_w = cell_w * 2
        box_h = cell_h * 2

        hot_zone = (
            int(max(0, center_x - box_w // 2)),
            int(max(0, center_y - box_h // 2)),
            int(min(box_w, orig_w)),
            int(min(box_h, orig_h)),
        )

        return ThermalStats(
            max_temp=float(max_val),
            min_temp=float(min_val),
            hot_zone=hot_zone,
            grid=grid,
        )