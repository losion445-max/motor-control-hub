import { useWsHub } from './hooks/useWsHub'; // предполагаемый путь

import React from 'react';
import type { SystemStatus } from 'src/domain/types';
import type { WsCommand } from 'src/domain/wstypes';

interface FieldProps {
  // Типизируем функцию отправки, принимающую наши команды
  sendCommand: (cmd: WsCommand) => void;
  // Статус может быть null, пока идет первое соединение
  status: SystemStatus | null;
}

export const Field: React.FC<FieldProps> = ({ sendCommand, status }) => {

  const handleFieldClick = (e: React.MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;

    sendCommand({
      type: 'MOVE',
      payload: { x, y, speed: 100 }
    });
  };

  return (
    <div
      onClick={handleFieldClick}
      style={{
        width: '500px',
        height: '500px',
        background: '#eee',
        position: 'relative',
        cursor: 'crosshair'
      }}
    >
      {/* ВНИМАНИЕ: Если статус null, мы ничего не рисуем. 
         Если статус есть, используем его данные.
      */}
      {status && (
        <div style={{
          position: 'absolute',
          // Добавляем 'px' как строку, так как style принимает string
          left: `${status.position.x}px`,
          top: `${status.position.y}px`,
          width: '10px',
          height: '10px',
          background: 'red',
          borderRadius: '50%',
          // Опционально: центрируем точку относительно координат
          transform: 'translate(-50%, -50%)',
          transition: 'all 0.017s linear' // Плавность в UI
        }} />
      )}
    </div>
  );
};

export default function App() {
  const { status, isLive, sendCommand } = useWsHub();

  return (
    <div style={{ padding: '20px', fontFamily: 'monospace' }}>
      <Field
        status={status}
        sendCommand={sendCommand}
      />
      <h1>System Control</h1>

      {/* Статус соединения */}
      <div style={{ background: isLive ? '#dfd' : '#fdd', padding: '5px' }}>
        Status: {isLive ? 'LIVE' : 'OFFLINE'}
      </div>

      {/* Координаты и управление */}
      {status && (
        <div style={{ marginTop: '20px' }}>
          <h3>Position: X: {status.position.x.toFixed(2)} | Y: {status.position.y.toFixed(2)}</h3>

          <div style={{ display: 'flex', gap: '10px' }}>
            <button onClick={() => sendCommand({ type: 'STOP' })}>
              STOP
            </button>
            <button onClick={() => sendCommand({
              type: 'MOVE',
              payload: { x: 100, y: 100, speed: 50 }
            })}>
              Move to 100,100
            </button>
          </div>

          <div style={{ marginTop: '20px' }}>
            <h4>Motors:</h4>
            <ul>
              {status.motors.map(m => (
                <li key={m.motor_id}>
                  Motor {m.motor_id}: {m.current_steps} steps |
                  {m.enabled ? ' ✅' : ' ❌'}
                </li>
              ))}
            </ul>
          </div>
        </div>
      )}

      {/* Секция логов или ошибок */}
      {!status && <p>Waiting for connection...</p>}
    </div>

  );
}