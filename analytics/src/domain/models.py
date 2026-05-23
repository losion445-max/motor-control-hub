from dataclasses import asdict, dataclass
import json
import numpy as np

@dataclass
class ThermalStats:
    max_temp: float
    min_temp: float
    hot_zone: tuple  # (x, y, w, h)
    grid: np.ndarray # Матрица 8x6
    
    def to_json(self):
            # Превращаем всё в словарь
            data = asdict(self)
            
            # NumPy массивы (grid) нужно превратить в обычные списки
            if isinstance(data['grid'], np.ndarray):
                data['grid'] = data['grid'].tolist()
                
            return json.dumps(data)