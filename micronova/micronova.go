package micronova

import (
	"micronova2mqtt/files"
	"micronova2mqtt/mqtt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	durationActive     = 30 * time.Second
	durationChecking   = time.Minute
	maxInactiveCycles  = 180
	https              = "https://"
	acceptHeader       = "application/json, text/javascript, */*; q=0.01"
	contentTypeHeader  = "application/json"
	brandIdHeader      = "1"
	requestTimeout     = 10 * time.Second
	stateInactive      = 0
	stateActive        = 1
	stateOffline       = 2
	stateNotResponding = 3
)

type valueDescr struct {
	Value       int    `json:"value"`
	Description string `json:"description"`
}

type parameter struct {
	regKey     string
	title      string
	offset     int
	mask       int
	minimum    int
	maximum    int
	step       int
	value      int
	valueDescr []valueDescr
	formula    string
	realValue  float64
	format     string
	text       string
}

type Queue interface {
	Publish(topic, value string)
	Subscribe(topic string, handler mqtt.SubscriberFunc)
}

var offsets []int
var parameters []parameter

var state int
var updateMqtt bool

var deviceName string
var deviceSerial string
var deviceModel string

var dm *files.DataManager
var customerCode string
var apiDomain string
var queue Queue

var powerOn = setPowerOn
var powerOff = setPowerOff
var actions ActionsHolder

func Run(dataManager *files.DataManager, customercode string, apidomain string, queuefunc Queue) {
	dm = dataManager
	customerCode = customercode
	apiDomain = apidomain
	queue = queuefunc

	if dm.Config.Micronova.Power.On != "" {
		powerOn = dm.Config.Micronova.Power.On
	}
	if dm.Config.Micronova.Power.Off != "" {
		powerOff = dm.Config.Micronova.Power.Off
	}

	actions.getCalculated = getCalculated
	actions.setParameterByRegKey = SetParameterByRegKey
	actions.parameters = &parameters

	for _, action := range dm.Config.Micronova.Actions {
		setValues := ""
		for _, setValue := range action.SetValues {
			setValues += setValue.SetKey + "=" + strconv.Itoa(setValue.Value) + " "
		}
		log.Info().Msgf("Action: %s >= %.1f & <=%.1f -->  %s", action.Trigger.GetKey, action.Trigger.MinimumValue, action.Trigger.MaximumValue, setValues)
	}

	for {
		// Signup - Register app UUID
		if dm.Session.UUID == "" {
			signup()
		}

		tokenExp, err := getJwtExpiration(dm.Session.Token)
		if err != nil {
			tokenExp = time.Now().Add(-24 * time.Hour) // set token to yesterday (expired)
		}
		log.Debug().Msgf("Token Expiration: %s", tokenExp)

		refreshTokenExp, err := getJwtExpiration(dm.Session.RefreshToken)
		if err != nil {
			refreshTokenExp = time.Now().Add(-24 * time.Hour) // set token to yesterday (expired)
		}
		log.Debug().Msgf("RefreshToken Expiration: %s", refreshTokenExp)

		// Login if RefreshToken expired (or empty)
		if refreshTokenExp.Before(time.Now()) {
			login()
		}

		// Update Token if JWT expired
		if tokenExp.Before(time.Now()) {
			err = updateToken()
			if err != nil {
				log.Error().Err(err).Msg("JWT Token Error - Tokens cleared")
				continue
			}
		}

		// Get ProductId and DeviceId (of first Device)
		if dm.Session.ProductId == "" || dm.Session.DeviceId == "" {
			deviceList()
		}

		// Check if device is online/offline
		if (state != stateActive && state != stateInactive) || deviceName == "" {
			deviceInfo()
		}

		if state == stateActive || state == stateInactive {
			// Get list of offsets if no RegKeys in config
			if len(dm.Config.Micronova.RegKeys) == 0 && len(offsets) == 0 {
				readDeviceBuffer()
			}

			// Get list of parameters, based on RegKeys or offsets
			if len(parameters) == 0 {
				deviceRegisters()

				topic := dm.Config.Mqtt.BaseTopic + "/" + deviceName + "/command/+"
				queue.Subscribe(topic, receiveCommand)

				if dm.Config.HomeAssistant.Discovery {
					// Register with Home Assistant Discovery
					discoveryConfig := createHAConfig(queue)
					discoveryConfig.Setup()
				}
			}

			// Read the device parameters
			readDevice()

			// run actions
			if state == stateActive {
				actions.Process(dm.Config.Micronova.Actions)
			}
		}

		// Publish the device parameters and values on MQTT Topic
		publishParameters()

		// Wait short time (30 sec) if device active or amount of cycles if inactive
		for range maxInactiveCycles {
			time.Sleep(durationActive)
			if isActive() || updateMqtt || state == stateNotResponding {
				updateMqtt = false
				break
			}
		}
	}
}
