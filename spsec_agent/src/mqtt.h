#pragma once
#include <Arduino.h>
#include <ESP8266WiFi.h>
#include <PubSubClient.h>

class MQTTClass {
private:
    WiFiClient wifiClient;
    PubSubClient mqtt;

    String server;
    int port;
    String clientId;

public:
    MQTTClass(const String& serverAddr, int serverPort)
        : server(serverAddr), port(serverPort), mqtt(wifiClient)
    {
        clientId = "esp8266-" + String(ESP.getChipId(), HEX);
    }

    void begin() {
        mqtt.setServer(server.c_str(), port);
    }

    void loop() {
        if (!mqtt.connected()) {
            reconnect();
        }
        mqtt.loop();
    }

    bool connected() {
        return mqtt.connected();
    }
    

    void reconnect() {
        if (mqtt.connected()) return;

        Serial.print("MQTT connecting to ");
        Serial.println(server);

        while (!mqtt.connected()) {
            if (mqtt.connect(clientId.c_str())) {
                Serial.println("MQTT connected");
            } else {
                Serial.print("Failed, rc=");
                Serial.print(mqtt.state());
                Serial.println(" retry in 1s");
                delay(1000);
            }
        }
    }

    void publish(const String& topic, const String& payload) {
        if (!mqtt.connected()) reconnect();
        mqtt.publish(topic.c_str(), payload.c_str());
    }

    void subscribe(const String& topic, MQTT_CALLBACK_SIGNATURE) {
        mqtt.setCallback(callback);
        mqtt.subscribe(topic.c_str());
    }
};

class MqttWorker {
    private:
        MQTTClass &mqtt;
        String topic;
    
    public:
        MqttWorker(MQTTClass &mqttClient, const String &t)
            : mqtt(mqttClient), topic(t) {}
    
        void sendState(int value) {
            mqtt.publish(topic, String(value));
            Serial.print("MQTT published: ");
            Serial.println(value);
        }
    };
    