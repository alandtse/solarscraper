// SPDX-FileCopyrightText: Contributors
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gosimple/slug"
)

var client mqtt.Client

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

// device Data
type Device struct {
	Identifiers  string `json:"identifiers"`
	Manufacturer string `json:"manufacturer"`
	Name         string `json:"name"`
	Version      string `json:"sw_version"`
}
type Registration struct {
	StateTopic    string `json:"state_topic"`
	Unit          string `json:"unit_of_measurement"`
	Name          string `json:"name"`
	ValueTemplate string `json:"value_template"`
	DeviceClass   string `json:"device_class,omitempty"`
	StateClass    string `json:"state_class"`
	Device        Device `json:"device"`
	UniqueID      string `json:"unique_id"`
	Icon          string `json:"icon,omitempty"`
}

func publish() {
	var reading SolarOSReading
	err := json.Unmarshal(jsonResponse, &reading)
	if err != nil {
		return
	}
	var lifemeter, _ = strconv.ParseFloat(reading.LifeMeter, 32)
	if reading.InstantaneousPower == "" || lifemeter < 2 {
		return
	}
	// fmt.Println(fmt.Printf("Reading %+v", reading))

	topic := "power/powerStats"
	payload, err := json.Marshal(&reading)
	if client == nil {
		login()
	}
	if err == nil {
		fmt.Println(fmt.Printf("%s:Publishing payload: %s", topic, string(payload)))
		token := client.Publish(topic, 0, false, string(payload))
		token.Wait()
		time.Sleep(time.Second)
	}
}

func publish_discovery() {
	// publish autodiscovery sensor info
	// https://www.home-assistant.io/docs/mqtt/discovery/
	address := slug.Make(config.MQTT.Address)
	solarosDevice := Device{
		Identifiers:  address,
		Manufacturer: "SolarOS",
		Name:         "SolarOS",
		Version:      "1",
	}
	RegistrationArray := [5]Registration{
		{
			StateTopic:    "power/powerStats",
			Unit:          "kW",
			Name:          "Current Generation",
			ValueTemplate: "{{ value_json.instant_power | float}}",
			DeviceClass:   "power",
			StateClass:    "measurement",
			Device:        solarosDevice,
			UniqueID:      fmt.Sprintf("solaros_%s_current", address),
		},
		{
			StateTopic:    "power/powerStats",
			Unit:          "kWh",
			Name:          "Lifetime Generated",
			ValueTemplate: "{{ value_json.life_meter | float }}",
			DeviceClass:   "energy",
			StateClass:    "total_increasing",
			Device:        solarosDevice,
			UniqueID:      fmt.Sprintf("solaros_%s_lifetime", address),
		},
		{
			StateTopic:    "power/powerStats",
			Unit:          "trees",
			Name:          "Trees Saved",
			ValueTemplate: "{{ value_json.trees_saved | float }}",
			StateClass:    "total_increasing",
			Device:        solarosDevice,
			UniqueID:      fmt.Sprintf("solaros_%s_trees", address),
			Icon:          "mdi:pine-tree",
		},
		{
			StateTopic:    "power/powerStats",
			Unit:          "lbs",
			Name:          "GHG Offset",
			ValueTemplate: "{{ value_json.oil_offset | float }}",
			StateClass:    "total_increasing",
			Device:        solarosDevice,
			UniqueID:      fmt.Sprintf("solaros_%s_oil_offset", address),
			Icon:          "mdi:oil",
		},
		{
			StateTopic:    "power/powerStats",
			Unit:          "hours",
			Name:          "Houses Offset",
			ValueTemplate: "{{ value_json.co2_offset | float }}",
			StateClass:    "total_increasing",
			Device:        solarosDevice,
			UniqueID:      fmt.Sprintf("solaros_%s_co2_offset", address),
			Icon:          "mdi:molecule-co2",
		},
	}

	for index, registration := range RegistrationArray {
		topic := fmt.Sprintf("%s/sensor/%s/config",
			config.MQTT.AutoDiscovery, registration.UniqueID)
		payload, err := json.Marshal(&registration)
		if err == nil {
			fmt.Println(fmt.Printf("%s: Registering sensor %d -> %s", topic, index, string(payload)))
			token := client.Publish(topic, 0, true, string(payload))
			token.Wait()
			time.Sleep(time.Second)
		}
	}

}

func login() {
	var broker = config.MQTT.Host
	var port = config.MQTT.Port
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%s", broker, port))
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername(config.MQTT.User)
	opts.SetPassword(config.MQTT.Pass)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	// opts.OnConnectionLost = connectLostHandler
	// opts.SetPingTimeout(10 * time.Second)
	opts.SetKeepAlive(10 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)
	// opts.SetReconnectingHandler(func(c mqtt.Client, options *mqtt.ClientOptions) {
	// 	fmt.Println("...... mqtt reconnecting ......")
	// })
	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if config.MQTT.AutoDiscovery != "" {
		publish_discovery()
	}
}
