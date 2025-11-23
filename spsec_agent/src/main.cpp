#include <Arduino.h>
#include "secrets.h"
#include "wifi.h"
#include "sensor.h"
#include "worker.h"

volatile int lastState = -1;
volatile bool stateChanged = false;

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
    door.calibrate();
    connection.connect();
    sender.lastSend = millis();
}

//
// ===== Loop =====
void loop() {
    connection.run();
    connection.keepAlive();



    if (stateChanged) {
        worker.sendAsync(lastState);
        stateChanged = false;
    }

    door.run();  // датчик всегда опрашивается

    sender.run();  // периодическая отправка
    worker.run();  // асинхронная отправка
}
