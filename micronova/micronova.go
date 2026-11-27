package micronova

import (
	"micronova2mqtt/appdata"
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

var dm *appdata.DataManager
var customerCode string
var apiDomain string
var queue Queue

var powerOn = setPowerOn
var powerOff = setPowerOff
var actions ActionsHolder

func checkSession() error {
	// Signup - Register app UUID
	if dm.Session.UUID == "" {
		if err := signup(); err != nil {
			log.Error().Err(err).Msg("Signup Error")
			return err
		}
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
		if err := login(); err != nil {
			log.Error().Err(err).Msg("Login Error")
			return err
		}
	}

	// Update Token if JWT expired
	if tokenExp.Before(time.Now()) {
		if err := updateToken(); err != nil {
			log.Error().Err(err).Msg("JWT Token Update Error")
			return err
		}
	}

	// Get ProductId and DeviceId (of first Device)
	if dm.Session.ProductId == "" || dm.Session.DeviceId == "" {
		if err := deviceList(); err != nil {
			log.Error().Err(err)
			return err
		}
	}

	return nil
}

func Run(dataManager *appdata.DataManager, customercode string, apidomain string, queuefunc Queue) {
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
		if err := checkSession(); err != nil {
			time.Sleep(durationActive)
			continue
		}

		// Get Device Info and Check if device is online/offline
		if (state != stateActive && state != stateInactive) || deviceName == "" {
			if err := deviceInfo(); err != nil {
				time.Sleep(durationActive)
				continue
			}
		}

		if state == stateActive || state == stateInactive {
			// Get list of offsets if no RegKeys in config
			if len(dm.Config.Micronova.RegKeys) == 0 && len(offsets) == 0 {
				if err := readDeviceBuffer(); err != nil {
					time.Sleep(durationActive)
					continue
				}
			}

			// Get list of parameters, based on RegKeys or offsets
			if len(parameters) == 0 {
				if err := deviceRegisters(); err != nil {
					time.Sleep(durationActive)
					continue
				}

				topic := dm.Config.Mqtt.BaseTopic + "/" + deviceName + "/command/+"
				queue.Subscribe(topic, receiveCommand)

				if dm.Config.HomeAssistant.Discovery {
					// Register with Home Assistant Discovery
					discoveryConfig := createHAConfig(queue)
					discoveryConfig.Setup()
				}
			}

			// Read the device parameters
			if err := readDevice(); err != nil {
				time.Sleep(durationActive)
				continue
			}

			// run actions
			if state == stateActive {
				actions.Process(dm.Config.Micronova.Actions)
			}

			// Publish the device parameters and values on MQTT Topic
			publishParameters()
		}

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
