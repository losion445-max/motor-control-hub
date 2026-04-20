# Содержимое файла README.md

# 🚀 Motor Control Hub

Система управления моторами через веб-интерфейс. Построена на базе микросервисной архитектуры в Docker.

## 🏗 Архитектура

- **Frontend**: React + Vite + Tailwind CSS (раздается через Nginx).
- **Backend**: Go (Chi router) — проксирует команды на ESP32.

## 🛠 Быстрый старт (Docker)

Для запуска всей инфраструктуры (Frontend + Backend) одной командой:

```bash
docker compose up --build
```

## Сделано

### Инициализация
Го бекенд + vite + rect + ts через nginx прокси на 80 порт
