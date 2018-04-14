# mqtt2influx
Subscribe to an mqtt topic and push the received values to an influxdb 

## Usecase
 
I have some ESP8266 boards with an BME280 sensor in my house
to measure mainly temperature and humidty. These boards run the  [esper](https://github.com/esper-hub/esper) firmware
and publish the measurements of the sensor via MQTT. This is perfectly fine to integrate them into [Home Assistant](https://github.com/home-assistant/home-assistant), but not
perfect for longer retention rates and to fulfill my secret passion for fancy graphs.

Therefore I wrote this tool, which subscribes to the defined MQTT topics, and pushes every received value into an InfluxDB. This way
it can easily be used as a backend for [Grafana](https://github.com/grafana/grafana).

