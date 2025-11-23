#pragma once

#include <ESP8266WiFi.h>
#include "secrets.h"

extern volatile int lastState;

class Worker {
public:
    enum State {IDLE, CONNECTING, SENDING, READING, DONE};
    State state = IDLE;

    int payload = 0;
    unsigned long sendStartTime = 0;
    unsigned long timeout = 100000; // ms
    WiFiClient client;
    int port = 1312;

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

const int sendInterval = 5000; // ms

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
 // WORKER_H