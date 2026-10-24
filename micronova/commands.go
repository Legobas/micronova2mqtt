package micronova

import (
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

const (
	setPowerOn         = "on"
	setPowerOff        = "off"
	statusManagedGet   = 232
	statusManagedOn    = 85
	statusManagedOff   = 170
	statusManagedMask  = 65535
	powerCommand       = "power"
)

func receiveCommand(topic string, value string) {
	lastSlashIdx := strings.LastIndex(topic, "/")

	if lastSlashIdx == -1 {
		log.Warn().Msgf("Incorrect Topic: %s", topic)
		return
	}

	command := topic[lastSlashIdx+1:]

	log.Debug().Msgf("Command: %s=%s", command, value)

	if command == powerCommand {
		// Switch device on/off
		SetPower(value)
	} else {
		SetParameterByTitle(command, value)
	}
	SetUpdateMqtt(true)
}

func SetPower(command string) {
	topicStart := dm.Config.Mqtt.BaseTopic + "/" + deviceName + "/"

	switch command {
	case powerOn:
		writeDevice(statusManagedGet, statusManagedOn, statusManagedMask)
		actions.reset()
		log.Info().Msg("Power On")
		queue.Publish(topicStart+"power", setPowerOn)
	case powerOff:
		writeDevice(statusManagedGet, statusManagedOff, statusManagedMask)
		log.Info().Msg("Power Off")
		queue.Publish(topicStart+"power", setPowerOff)
	default:
		log.Warn().Msgf("Invalid power value: %s", command)
	}
}

func SetParameterByTitle(title string, val string) {
	for _, param := range parameters {
		if title == param.title {
			value, err := strconv.Atoi(val)
			if err != nil {
				log.Error().Msgf("Incorrect parameter: %v", err)
				return
			}
			setParameter(param, value)
			break
		}
	}
}

func SetParameterByRegKey(key string, value int) {
	for _, par := range parameters {
		if par.regKey == key {
			setParameter(par, value)
			break
		}
	}
}

func setParameter(param parameter, value int) {
	if value > param.maximum {
		log.Error().Msgf("%s value %d is greater than maximum (%d)", param.title, value, param.maximum)
		return
	}
	if value < param.minimum {
		log.Error().Msgf("%s value %d is lower than minimum (%d)", param.title, value, param.minimum)
		return
	}
	log.Info().Msgf("MQTT command: Set Parameter %s to %d", param.title, value)

	// Send the device parameter to Micronova
	writeDevice(param.offset, value, param.mask)
}

func SetUpdateMqtt(update bool) {
	updateMqtt = update
}
