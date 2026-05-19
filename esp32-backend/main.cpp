/**
 * ============================================================
 * ESP32 Motor Controller — Clean Architecture Monolith (UDP Control)
 * Core Mechanical Logic + Hardware ISR Drivers + HTTP (Discovery) + UDP (Streaming)
 * ============================================================
 */

#include <Arduino.h>
#include <WiFi.h>
#include <WiFiUdp.h>
#include <WebServer.h>

// ============================================================
//  1. DOMAIN LAYER (Конфигурация, Спецификации и Состояние)
// ============================================================

namespace Domain {
    const float MAX_HZ = 300000.0f; // 300 кГц — жесткий лимит прерываний ESP32
    const float MIN_HZ = 1.0f;      // Минимальная частота таймера

    // Спецификация исполнительного механизма
    struct DeviceConfig {
        const uint8_t motor_id       = 1;
        const float pulley_diameter  = 90.0f;
        const long steps_per_rev     = 10000L;
        const float distance_per_rev = (pulley_diameter / 1000.0f) * PI;

        // Дифференциальные пины
        const uint8_t step_plus      = 26;
        const uint8_t step_minus     = 32;
        const uint8_t dir_plus       = 25;
        const uint8_t dir_minus      = 33;
    };

    // Атомарное операционное состояние мотора
    struct EngineState {
        volatile long target_steps   = 0;
        volatile long current_steps  = 0;
        volatile bool is_enabled     = false;
        volatile bool is_infinite    = false;
        volatile bool step_phase     = false; // false=LOW, true=HIGH
        volatile float current_hz    = 100.0f; // Храним честную частоту в Гц
        volatile uint8_t last_seq    = 0;
    };

    // Строго 12 байт, Little Endian — один в один как StreamPacket в Go
    #pragma pack(push, 1)
    struct UdpCommandPacket {
        uint8_t  magic;       // 0x5A
        uint8_t  motor_id;    // 1..4
        uint8_t  state_cmd;   // 0 - IDLE, 1 - STREAMING, 2 - ESTOP
        uint8_t  seq;         // Порядковый номер пакета
        int32_t  target_step; // 4 байта
        float    linear_speed;// Переименовали speed_rps -> скорость из Go (мм/сек). Размер те же 4 байта float.
    };
    #pragma pack(pop)

    class MotionCalculators {
    public:
        static float VelocityToHz(float mm_per_second, const DeviceConfig& spec) {
            // Переводим скорость из мм/сек в м/сек, так как в формуле pulley_diameter / 1000.0f
            float meters_per_second = fabsf(mm_per_second) / 1000.0f;
            return meters_per_second * (float)spec.steps_per_rev / spec.distance_per_rev;
        }
    };
}

static const Domain::DeviceConfig  g_device_spec;
static Domain::EngineState         g_engine;

// ============================================================
//  2. INFRASTRUCTURE LAYER (Hardware Drivers, ISR & Network Sockets)
// ============================================================

namespace Infrastructure {
    static hw_timer_t* g_hw_timer = nullptr;
    static portMUX_TYPE g_timer_mux = portMUX_INITIALIZER_UNLOCKED;

    // Двухфазный ISR (Вызывается на HIGH и на LOW)
    void IRAM_ATTR MotorInterruptHandler() {
        portENTER_CRITICAL_ISR(&g_timer_mux);

        if (!g_engine.is_enabled) {
            portEXIT_CRITICAL_ISR(&g_timer_mux);
            return;
        }

        if (!g_engine.is_infinite && g_engine.current_steps >= llabs(g_engine.target_steps)) {
            g_engine.is_enabled = false;
            digitalWrite(g_device_spec.step_plus,  LOW);
            digitalWrite(g_device_spec.step_minus, HIGH);
            portEXIT_CRITICAL_ISR(&g_timer_mux);
            return;
        }

        g_engine.step_phase = !g_engine.step_phase;

        if (g_engine.step_phase) {
            digitalWrite(g_device_spec.step_plus,  HIGH);
            digitalWrite(g_device_spec.step_minus, LOW);
        } else {
            digitalWrite(g_device_spec.step_plus,  LOW);
            digitalWrite(g_device_spec.step_minus, HIGH);
            if (!g_engine.is_infinite) {
                g_engine.current_steps++;
            }
        }

        portEXIT_CRITICAL_ISR(&g_timer_mux);
    }

    class LowLevelDriver {
    public:
        static void InitPins(const Domain::DeviceConfig& spec) {
            pinMode(spec.step_plus,  OUTPUT);
            pinMode(spec.step_minus, OUTPUT);
            pinMode(spec.dir_plus,   OUTPUT);
            pinMode(spec.dir_minus,  OUTPUT);

            digitalWrite(spec.step_plus,  LOW);
            digitalWrite(spec.step_minus, HIGH);
            digitalWrite(spec.dir_plus,   LOW);
            digitalWrite(spec.dir_minus,  HIGH);
        }

        static void SetDirection(bool forward, const Domain::DeviceConfig& spec) {
            if (forward) {
                digitalWrite(spec.dir_plus,  HIGH);
                digitalWrite(spec.dir_minus, LOW);
            } else {
                digitalWrite(spec.dir_plus,  LOW);
                digitalWrite(spec.dir_minus, HIGH);
            }
            delayMicroseconds(10); // Задержка DIR перед первого шага STEP
        }

        static void ApplyTimerFrequency(float hz) {
            if (hz < Domain::MIN_HZ) hz = Domain::MIN_HZ;
            if (hz > Domain::MAX_HZ) hz = Domain::MAX_HZ;

            uint64_t alarm_value = (uint64_t)(500000.0f / hz); // 2 прерывания на импульс
            if (alarm_value < 1) alarm_value = 1;

            if (g_hw_timer != nullptr) {
                timerAlarmDisable(g_hw_timer);
                timerDetachInterrupt(g_hw_timer);
                timerEnd(g_hw_timer);
                g_hw_timer = nullptr;
            }

            g_hw_timer = timerBegin(0, 80, true); // 1 Гц базовая частота тиков
            timerAttachInterrupt(g_hw_timer, &MotorInterruptHandler, true);
            timerAlarmWrite(g_hw_timer, alarm_value, true);
            timerAlarmEnable(g_hw_timer);
        }

        static void ForceStop(const Domain::DeviceConfig& spec) {
            g_engine.is_enabled = false;
            g_engine.is_infinite = false;
            
            portENTER_CRITICAL(&g_timer_mux);
            g_engine.target_steps = g_engine.current_steps;
            portEXIT_CRITICAL(&g_timer_mux);

            digitalWrite(spec.step_plus,  LOW);
            digitalWrite(spec.step_minus, HIGH);
        }
    };
}

// ============================================================
//  3. USECASE / BUSINESS LAYER (Управление стейтом привода)
// ============================================================

namespace Usecase {
    class MotorController {
    public:
        static void HandleTargetState(long steps, float speed_hz, uint8_t state_cmd) {
            // Если прилетел стоп или ESTOP (state_cmd == 0 или 2)
            if (state_cmd == 0 || state_cmd == 2) {
                Infrastructure::LowLevelDriver::ForceStop(g_device_spec);
                return;
            }

            if (steps == 0) return;
            if (speed_hz < Domain::MIN_HZ) speed_hz = Domain::MIN_HZ;
            if (speed_hz > Domain::MAX_HZ) speed_hz = Domain::MAX_HZ;

            // Если целевые шаги или скорость изменились — перенастраиваем таймер
            if (g_engine.target_steps != llabs(steps) || g_engine.current_hz != speed_hz || !g_engine.is_enabled) {
                portENTER_CRITICAL(&Infrastructure::g_timer_mux);
                g_engine.target_steps  = llabs(steps);
                g_engine.current_steps = 0;
                g_engine.step_phase    = false;
                g_engine.current_hz    = speed_hz;
                g_engine.is_infinite   = false;
                portEXIT_CRITICAL(&Infrastructure::g_timer_mux);

                Infrastructure::LowLevelDriver::SetDirection(steps > 0, g_device_spec);
                Infrastructure::LowLevelDriver::ApplyTimerFrequency(speed_hz);
                g_engine.is_enabled = true;
            }
        }
    };
}

// ============================================================
//  4. PRESENTATION LAYER (HTTP Discovery & UDP Packet Receiver)
// ============================================================

namespace Presentation {
    const char* NETWORK_SSID   = "svinki";
    const char* NETWORK_PASS   = "12340000";
    const uint16_t UDP_PORT    = 8888;
    
    WebServer g_http_server(80);
    WiFiUDP   g_udp_socket;

    class NetworkManager {
    public:
        static void EstablishConnection() {
            Serial.printf("[NET] Connecting to Wi-Fi: \"%s\"...\n", NETWORK_SSID);
            WiFi.mode(WIFI_STA);
            WiFi.begin(NETWORK_SSID, NETWORK_PASS);

            uint8_t attempts = 0;
            while (WiFi.status() != WL_CONNECTED && attempts < 20) {
                delay(500);
                Serial.print('.');
                attempts++;
            }

            if (WiFi.status() == WL_CONNECTED) {
                Serial.printf("\n[NET] Wi-Fi Connected. IP: %s\n", WiFi.localIP().toString().c_str());
                g_udp_socket.begin(UDP_PORT);
                Serial.printf("[NET] UDP Receiver bound to port %d\n", UDP_PORT);
            } else {
                Serial.println("\n[NET] Fatal: Connection Timeout.");
            }
        }

        static void MountDiscoveryRoute() {
            g_http_server.on("/config", HTTP_GET, []() {
                String data = "{";
                data += "\"motor_id\":"      + String(g_device_spec.motor_id);
                data += ",\"step_plus\":"    + String(g_device_spec.step_plus);
                data += ",\"step_minus\":"   + String(g_device_spec.step_minus);
                data += ",\"dir_plus\":"     + String(g_device_spec.dir_plus);
                data += ",\"dir_minus\":"    + String(g_device_spec.dir_minus);
                data += ",\"steps_per_rev\":" + String(g_device_spec.steps_per_rev);
                data += ",\"pulley_mm\":"     + String(g_device_spec.pulley_diameter, 1);
                data += "}";
                
                String json = "{\"status\":\"ok\",\"message\":\"discovery\",\"data\":" + data + "}";
                g_http_server.send(200, "application/json", json);
            });

            g_http_server.begin();
            Serial.println("[NET] HTTP Discovery Router ready on port 80");
        }

        static void PollUdpChannels() {
            int packet_size = g_udp_socket.parsePacket();
            if (packet_size == 0) return;

            if (packet_size == sizeof(Domain::UdpCommandPacket)) {
                Domain::UdpCommandPacket packet;
                g_udp_socket.read((char*)&packet, sizeof(Domain::UdpCommandPacket));

                if (packet.magic != 0x5A || packet.motor_id != g_device_spec.motor_id) {
                    return;
                }

                if (packet.seq != g_engine.last_seq) {
                    g_engine.last_seq = packet.seq;

                    // ВАЖНО: Переводим входящую линейную скорость (мм/сек) в честную частоту (Гц)
                    float calculated_hz = Domain::MotionCalculators::VelocityToHz(packet.linear_speed, g_device_spec);

                    // Отдаем рассчитанную частоту шагов в обработчик
                    Usecase::MotorController::HandleTargetState(
                        packet.target_step, 
                        calculated_hz, 
                        packet.state_cmd
                    );
                }
            } else {
                g_udp_socket.flush();
            }
        }

        static void HandleRuntimeMaintenance() {
            g_http_server.handleClient();
            PollUdpChannels();

            if (WiFi.status() != WL_CONNECTED) {
                Serial.println("[NET] Wi-Fi lost. Reconnecting...");
                EstablishConnection();
            }
        }
    };
}

// ============================================================
//  5. SYSTEM KERNEL ENTRYPOINTS
// ============================================================

void setup() {
    Serial.begin(115200);
    delay(500);
    Serial.printf("\n[KERNEL] Starting Node Topology. Motor Scope ID: %d\n", g_device_spec.motor_id);

    Infrastructure::LowLevelDriver::InitPins(g_device_spec);
    Presentation::NetworkManager::EstablishConnection();
    Presentation::NetworkManager::MountDiscoveryRoute();

    Serial.println("[KERNEL] Boot complete. Realtime scheduling active.");
}

void loop() {
    Presentation::NetworkManager::HandleRuntimeMaintenance();

    static unsigned long checkpoint = 0;
    if (millis() - checkpoint > 5000) {
        Serial.printf("[IO] Run=%d | TargetSteps=%ld | CurrentSteps=%ld | TargetHz=%.1f | RSSI=%d\n",
                      g_engine.is_enabled, g_engine.target_steps, 
                      g_engine.current_steps, g_engine.current_hz, WiFi.RSSI());
        checkpoint = millis();
    }

    delay(1);
}