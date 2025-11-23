#pragma once
#include <Arduino.h>

extern volatile int lastState;
extern volatile bool stateChanged;

class DoorSensor {
public:
    // === Параметры ===
    int baseline = 0;
    int threshold = -50;
    unsigned long debounceTime = 50; // ms
    unsigned long lastChangeTime = 0;
    int lastStableState = -1;

    unsigned long sensorInterval = 10;     // <-- Период опроса датчика
    unsigned long lastSensorRun = 0;

    enum SensorState {
        UNCALIBRATED,
        CALIBRATING,
        READY
    };
    SensorState sensorState = UNCALIBRATED;

    long calibrationSum = 0;
    int calibrationCount = 0;
    unsigned long lastCalibrationTime = 0;

    // === Запуск калибровки ===
    void calibrate() {
        if (sensorState == UNCALIBRATED) {
            Serial.println("Starting calibration...");
            sensorState = CALIBRATING;
            calibrationSum = 0;
            calibrationCount = 0;
            lastCalibrationTime = millis();
        }
    }

    // === Чтение сенсора ===
    int readRaw() {
        return analogRead(A0);
    }

    
    // === Главный метод ===
    void run() {
        unsigned long now = millis();

        // --- Интервальное исполнение ---
        if (now - lastSensorRun < sensorInterval) return;
        lastSensorRun = now;

        // --- Калибровка ---
        if (sensorState == CALIBRATING) {
            if (now - lastCalibrationTime >= 2) {
                calibrationSum += readRaw();
                calibrationCount++;
                lastCalibrationTime = now;

                if (calibrationCount >= 200) {
                    baseline = calibrationSum / 200;
                    Serial.print("Baseline = ");
                    Serial.println(baseline);
                    sensorState = READY;
                }
            }
            return;
        }

        if (sensorState != READY) return;
        // Serial.println("Sensor ready");
        // --- Основное чтение ---
        int currentReading = readRaw();
        int currentState = lastStableState;
        // Serial.printf("Door reading: %d\n", currentReading);

        if (currentReading > 800) {
            currentState = 0; // CLOSED
        } else if (currentReading > baseline + threshold) {
            currentState = 1; // OPEN
        }

        // --- Дребезг ---
        if (currentState != lastStableState) {
            lastChangeTime = now;
        }

        if ((now - lastChangeTime) > debounceTime) {

            // Serial.println(currentState == 1 ? "OPEN" : "CLOSED");
            // Serial.printf("Door current: %d last %d\n", currentState, lastStableState);/
            if (currentState != lastState) {
                lastState = currentState;
                stateChanged = true;

                Serial.print("Door changed: ");
                Serial.println(lastState == 1 ? "OPEN" : "CLOSED");
            }
        }

        lastStableState = currentState;
    }
};