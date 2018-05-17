package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/spf13/viper"
)

func MQTTMessageHandler(mqtt_client mqtt.Client, msg mqtt.Message, c client.Client, db string) {
	t := msg.Topic()
	p := string(msg.Payload())
	_, err := strconv.ParseFloat(p, 64)
	if err != nil {
		log.Printf("the payload on %s was not a float number", t)
		return
	}

	if !strings.Contains(t, "bme") {
		// TODO: Maybe use a regexp from the config file to
		// differentiate between good and bad topics
		return
	}

	topicParts := strings.Split(t, "/")
	id, sensor := topicParts[2], topicParts[4]

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  db,
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	fields := map[string]interface{}{
		"value":  p,
		"sensor": sensor,
	}

	s := []string{id, sensor}
	seriesName := strings.Join(s, "-")
	pt, err := client.NewPoint(seriesName, nil, fields, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := c.Write(bp); err != nil {
		log.Fatal(err)
	}
	log.Printf("Added %6s from the %11s sensor of the BME-%s", p, sensor, id)
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	influxConf := viper.Sub("influx")
	influxClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     influxConf.GetString("host"),
		Username: influxConf.GetString("user"),
		Password: influxConf.GetString("pass"),
	})

	log.Print("Connected to influxdb")
	if err != nil {
		log.Fatal(err)
	}

	mqttConf := viper.Sub("mqtt")
	opts := mqtt.NewClientOptions().AddBroker(mqttConf.GetString("broker")).SetUsername(mqttConf.GetString("user")).SetPassword(mqttConf.GetString("pass"))
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	wg := sync.WaitGroup{}
	wg.Add(1)

	if token := mqttClient.Subscribe(mqttConf.GetString("topic"), 0, func(mqttClient mqtt.Client, msg mqtt.Message) {
		MQTTMessageHandler(mqttClient, msg, influxClient, influxConf.GetString("db"))
	}); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	wg.Wait()
}
