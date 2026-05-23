from abc import ABC, abstractmethod

import numpy as np

from src.domain.models import ThermalStats

class ThermalEngine(ABC):
    @abstractmethod
    def analyze(self, frame: np.ndarray) -> ThermalStats:
        pass

class DataStorage(ABC):
    @abstractmethod
    def save_stats(self, stats: ThermalStats):
        pass

class VideoStreamer(ABC):
    @abstractmethod
    def push_frame(self, frame: np.ndarray):
        pass