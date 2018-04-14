package main

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/influxdata/influxdb/client/v2"
	"log"
	"sync"
	"time"
	"strings"
)

const (
	influx_host = "http://localhost:8086"
	influx_db   = "mqtt"
	influx_user = "influxuser"
	influx_pass = "supersecret"
	mqtt_broker = "tcp://mqttbroker:1883"
	mqtt_user   = "mqttuser"
	mqtt_pass   = "mqttpass"
	mqtt_topic  = "ibg10/esper/#"
)

func MQTTMessageHandler(mqtt_client mqtt.Client, msg mqtt.Message, c client.Client) {
	t := msg.Topic()
	p := string(msg.Payload())

	if strings.Contains(t, "bme"){

		topicParts := strings.Split(t, "/")
		id, sensor := topicParts[2], topicParts[4]
		log.Printf("Got %s for sensor %s on BME %s", p, sensor, id)
	}
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  influx_db,
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	fields := map[string]interface{}{
		"value": p,
	}

	pt, err := client.NewPoint(t, nil, fields, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := c.Write(bp); err != nil {
		log.Fatal(err)
	}
}

func main() {
	influx_client, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     influx_host,
		Username: influx_user,
		Password: influx_pass,
	})

	log.Print("Connected to influxdb")
	if err != nil {
		log.Fatal(err)
	}

	opts := mqtt.NewClientOptions().AddBroker(mqtt_broker).SetUsername(mqtt_user).SetPassword(mqtt_pass)
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	wg := sync.WaitGroup{}
	wg.Add(1)

	if token := mqttClient.Subscribe(mqtt_topic, 0, func(mqttClient mqtt.Client, msg mqtt.Message) {
		MQTTMessageHandler(mqttClient, msg, influx_client)
	}); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	wg.Wait()
}
