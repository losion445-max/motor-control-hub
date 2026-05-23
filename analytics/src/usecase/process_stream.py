# src/use_cases/process_stream.py
from src.domain.interfaces import DataStorage, ThermalEngine, VideoStreamer


class StreamProcessor:
    def __init__(self, engine: ThermalEngine, storage: DataStorage, streamer: VideoStreamer):
        self.engine = engine
        self.storage = storage
        self.streamer = streamer

    def execute(self, frame):
        stats = self.engine.analyze(frame)
        
        self.storage.save_stats(stats)

        self.streamer.push_frame(frame)