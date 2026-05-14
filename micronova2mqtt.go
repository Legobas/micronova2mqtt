package main

import (
	_ "embed"
	"fmt"
	"micronova2mqtt/files"
	"micronova2mqtt/micronova"
	"micronova2mqtt/mqtt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	appName  = "micronova2mqtt"
	powerKey = "Power"
)

var (
	//Version import - build with -ldflags "-X main.Version=${VERSION}"
	Version  string
	dataMgr  *files.DataManager
	mqttConn *mqtt.MqttConnection
)

func init() {
	// Setup logging
	out := zerolog.NewConsoleWriter()
	out.NoColor = true
	out.FormatLevel = func(i any) string {
		return strings.ToUpper(fmt.Sprintf("%-6s", i))
	}
	out.PartsExclude = []string{zerolog.TimestampFieldName, zerolog.CallerFieldName}
	log.Logger = log.Output(out)

	switch strings.ToLower(os.Getenv("LOGLEVEL")) {
	case "debug":
		log.Logger = log.With().Caller().Logger()
		log.Logger = log.Level(zerolog.DebugLevel)
	case "trace":
		log.Logger = log.With().Caller().Logger()
		log.Logger = log.Level(zerolog.TraceLevel)
	default:
		log.Logger = log.Level(zerolog.InfoLevel)
	}

	// Print Application with version
	log.Info().Msgf("%s %s", strings.ToUpper(appName), Version)
}

func receiveMqttMessage(key string, value string) {
	if key == powerKey {
		// Switch device on/off
		micronova.SetPower(value)
	} else {
		micronova.SetParameter(key, value)
	}
	micronova.SetUpdateMqtt(true)
	log.Debug().Msgf("MQTT message received: %s=%s", key, value)
}

func publishMqttMessage(category, key, value string, retain bool) {
	if mqttConn != nil {
		mqttConn.Publish(category, key, value, retain)
	}
}

func main() {
	// Initialize Config
	var err error
	dataMgr, err = files.NewData()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize config")
	}

	var mqttProperties mqtt.MqttProperties
	mqttProperties.ClientId = fmt.Sprintf("%s_%s", appName, fmt.Sprintf("%d", time.Now().UnixMicro()))
	mqttProperties.Url = dataMgr.Config.Mqtt.Url
	mqttProperties.User = dataMgr.Config.Mqtt.Username
	mqttProperties.Password = dataMgr.Config.Mqtt.Password
	mqttProperties.Qos = byte(dataMgr.Config.Mqtt.Qos)
	mqttProperties.Retain = dataMgr.Config.Mqtt.Retain
	mqttProperties.Receiver = receiveMqttMessage
	mqttConn, err = mqtt.NewMqttConnection(mqttProperties)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize MQTT")
	}

	customerCode, apiDomain, err := dataMgr.GetBrand(dataMgr.Config.Micronova.Brand)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get brand data")
	}

	micronova.Run(dataMgr, customerCode, apiDomain, publishMqttMessage)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	log.Info().Msgf("%s stopped", strings.ToUpper(appName))
}
