#include <Arduino.h>
#include <ESP8266WiFi.h>
#include <secrets.h>

int port = 1312;

const int sendInterval = 5000; // ms
const int doorPin = 5;

volatile int lastState = 2;
volatile bool stateChanged = false;

//
// ===== WIFI =====
class WIFI {
public:
    void connect() {
        WiFi.mode(WIFI_STA);
        WiFi.begin(wifiName.c_str(), wifiPass.c_str());
        Serial.print("Connecting WiFi");
        while (WiFi.status() != WL_CONNECTED) {
            delay(500);
            Serial.print(".");
        }
        Serial.println("\nWiFi connected!");
        Serial.print("IP: "); Serial.println(WiFi.localIP());
    }

    void keepAlive() {
        if (WiFi.status() != WL_CONNECTED) {
            Serial.println("WiFi lost. Reconnecting...");
            connect();
        }
    }
};

//
// ===== DoorSensor =====
class DoorSensor {
public:
    const unsigned long debounceTime = 50; // ms
    const unsigned long checkInterval = 10; // ms
    int lastStableState = HIGH;
    unsigned long lastChangeTime = 0;
    unsigned long lastCheckTime = 0;

    int readRaw() { return digitalRead(doorPin); }

    void run() {
        unsigned long now = millis();
        if (now - lastCheckTime >= checkInterval) {
            lastCheckTime = now;
            int current = readRaw();

            if (current != lastStableState) {
                if (now - lastChangeTime >= debounceTime) {
                    lastStableState = current;
                    lastChangeTime = now;
                    lastState = current;
                    stateChanged = true;

                    Serial.print("Door changed: ");
                    Serial.println(current == HIGH ? "OPEN" : "CLOSED");
                }
            }
        }
    }
};

//
// ===== Worker (полностью неблокирующий) =====
class Worker {
public:
    enum State {IDLE, CONNECTING, SENDING, READING, DONE};
    State state = IDLE;

    int payload = 0;
    unsigned long sendStartTime = 0;
    unsigned long timeout = 3000; // ms
    WiFiClient client;

    void sendAsync(int val) {
        if (state == IDLE) {
            payload = val;
            state = CONNECTING;
            sendStartTime = millis();
        }
    }

    void run() {
        switch(state) {
            case IDLE:
                return;

            case CONNECTING:
                if (client.connect(host.c_str(), port)) {
                    String data = String(payload);
                    String request = "POST / HTTP/1.1\r\n";
                    request += "Host: " + host + "\r\n";
                    request += "Content-Type: text/plain\r\n";
                    request += "Content-Length: " + String(data.length()) + "\r\n\r\n";
                    request += data;
                    client.print(request);
                    state = READING;
                    sendStartTime = millis();
                } else {
                    // не удалось подключиться, сбрасываем
                    Serial.printf("%d - ERR_CONNECT\n", payload);
                    client.stop();
                    state = IDLE;
                }
                break;

            case READING:
                // читаем ответ кусками, но не блокируем loop
                if (client.available()) {
                    String line = client.readStringUntil('\n');
                    line.trim();
                    if (line.startsWith("HTTP/1.1")) {
                        int code = line.substring(9,12).toInt();
                        Serial.printf("%d - %d\n", payload, code);
                        client.stop();
                        state = IDLE;
                    }
                }
                // таймаут
                if (millis() - sendStartTime > timeout) {
                    Serial.printf("%d - ERR_TIMEOUT\n", payload);
                    client.stop();
                    state = IDLE;
                }
                break;

            default:
                state = IDLE;
                break;
        }
    }
};

//
// ===== SendTimer =====
class SendTimer {
public:
    unsigned long lastSend = 0;
    Worker* worker;

    SendTimer(Worker* w) : worker(w) {}

    void run() {
        unsigned long now = millis();
        if (now - lastSend >= sendInterval) {
            worker->sendAsync(lastState);
            lastSend = now;
        }
    }
};

//
// ===== Objects =====
WIFI connection;
DoorSensor door;
Worker worker;
SendTimer sender(&worker);

//
// ===== Setup =====
void setup() {
    Serial.begin(115200);
    delay(500);
    pinMode(doorPin, INPUT_PULLUP);

    connection.connect();
    sender.lastSend = millis();
}

//
// ===== Loop =====
void loop() {
    connection.keepAlive();

    door.run();  // датчик всегда опрашивается

    if (stateChanged) {
        worker.sendAsync(lastState);
        stateChanged = false;
    }

    sender.run();  // периодическая отправка
    worker.run();  // асинхронная отправка
}
