#pragma once

#include <ESP8266WiFi.h>
#include "secrets.h"

extern volatile int lastState;

class WIFI {
private:
    enum WifiState {
        DISCONNECTED,
        CONNECTING,
        CONNECTED,
        RECONNECTING
    };
    WifiState state = DISCONNECTED;
    unsigned long lastAttempt = 0;

public:
void connect() {
    if (state == CONNECTED) return;

    WiFi.mode(WIFI_STA);
    WiFi.setSleepMode(WIFI_NONE_SLEEP);

    WiFi.disconnect();           // <-- ОБЯЗАТЕЛЬНО
    delay(200);                  // <-- ДА, НУЖНО

    WiFi.begin(wifiName.c_str(), wifiPass.c_str());

    Serial.println("Connecting WiFi");
    state = CONNECTING;
    lastAttempt = millis();
}



    void run() {
        if (state == CONNECTING) {
            if (WiFi.status() == WL_CONNECTED) {
                Serial.println("\nWiFi connected!");
                Serial.print("IP: "); Serial.println(WiFi.localIP());
                state = CONNECTED;
            } else if (millis() - lastAttempt > 500) {
                Serial.print(".");
                lastAttempt = millis();
            }
        }
    }
    
    

    void keepAlive() {
        if (WiFi.status() != WL_CONNECTED) {
            if (state != CONNECTING) {
                Serial.println("WiFi lost. Reconnecting...");
                state = CONNECTING;
    
                WiFi.disconnect();       // <-- НУЖНО
                delay(200);              // <-- НУЖНО
    
                WiFi.begin(wifiName.c_str(), wifiPass.c_str());
                lastAttempt = millis();
            }
        }
    }
    
    
};
