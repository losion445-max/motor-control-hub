import subprocess
import numpy as np
from src.domain.interfaces import VideoStreamer

class FFmpegStreamer(VideoStreamer):
    def __init__(self, rtmp_url: str, width: int, height: int):
        self.cmd = [
            'ffmpeg',
            '-y',
            '-f', 'rawvideo',
            '-vcodec', 'rawvideo',
            '-pix_fmt', 'gray',
            '-s', f'{width}x{height}',
            '-r', '20',
            '-i', '-',
            '-c:v', 'libx264',
            '-pix_fmt', 'yuv420p',
            '-preset', 'ultrafast',
            '-tune', 'zerolatency',
            '-f', 'flv',
            rtmp_url
        ]
        self.proc = subprocess.Popen(self.cmd, stdin=subprocess.PIPE, stderr=subprocess.PIPE)

    def push_frame(self, frame: np.ndarray):
        if frame.dtype != np.uint8:
            frame = frame.astype(np.uint8)
        
        try:
            self.proc.stdin.write(frame.tobytes())
            self.proc.stdin.flush()
        except BrokenPipeError:
            # Если ffmpeg упал, тут будет понятно почему
            err = self.proc.stderr.read()
            print(f"FFmpeg error: {err.decode()}", file=sys.stderr)
            raise

    def close(self):
        if self.proc.stdin:
            self.proc.stdin.close()
        self.proc.wait()