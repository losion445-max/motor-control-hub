from src.domain.interfaces import DataStorage
from src.domain.models import ThermalStats


class RedisStorage(DataStorage):
    def __init__(self, redis_client):
        self.client = redis_client
    def save_stats(self, stats: ThermalStats):
        self.client.set("cam1:stats", stats.to_json())