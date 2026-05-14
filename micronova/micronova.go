package micronova

import (
	"micronova2mqtt/files"
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
	setPowerOn         = "on"
	setPowerOff        = "off"
	statusManagedGet   = 232
	statusManagedOn    = 85
	statusManagedOff   = 170
	statusManagedMask  = 65535
	deviceOnline       = "Online"
	deviceOffline      = "Offline"
)

type valueDescr struct {
	Value       int    `json:"value"`
	Description string `json:"description"`
}

type parameter struct {
	regKey     string
	topicKey   string
	offset     int
	mask       int
	minimum    int
	maximum    int
	value      int
	valueDescr []valueDescr
	formula    string
	format     string
	text       string
}

type publishFunc func(category, key, value string, retain bool)

var offsets []int
var parameters []parameter

var state int
var updateMqtt bool
var deviceName string

var dm *files.DataManager
var customerCode string
var apiDomain string
var publisher publishFunc

func Run(dataManager *files.DataManager, customercode string, apidomain string, publishfunc publishFunc) {
	dm = dataManager
	customerCode = customercode
	apiDomain = apidomain
	publisher = publishfunc

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
			}

			// Read the device parameters
			readDevice()
		}

		// Publish the device parameters on MQTT Topic
		publishParameters()

		// Wait short time (20 sec) if device active or amount of cycles if inactive
		for range maxInactiveCycles {
			time.Sleep(durationActive)
			if isActive() || updateMqtt || state == stateNotResponding {
				updateMqtt = false
				break
			}
		}
	}
}

func SetPower(command string) {
	powerOn := setPowerOn
	powerOff := setPowerOff
	if dm.Config.Micronova.Power.On != "" {
		powerOn = dm.Config.Micronova.Power.On
	}
	if dm.Config.Micronova.Power.Off != "" {
		powerOff = dm.Config.Micronova.Power.Off
	}

	switch command {
	case powerOn:
		writeDevice(statusManagedGet, statusManagedOn, statusManagedMask)
		log.Info().Msg("Power On")
	case powerOff:
		writeDevice(statusManagedGet, statusManagedOff, statusManagedMask)
		log.Info().Msg("Power Off")
	default:
		log.Warn().Msgf("Invalid power value: %s", command)
	}
}

func SetParameter(key string, value string) {
	for _, par := range parameters {
		if key == par.topicKey {
			val, err := strconv.Atoi(value)
			if err != nil {
				log.Error().Msgf("Incorrect parameter: %v", err)
				return
			}
			if val > par.maximum {
				log.Error().Msgf("%s value %s is greater than maximum (%d)", key, value, par.maximum)
				return
			}
			if val < par.minimum {
				log.Error().Msgf("%s value %s is lower than minimum (%d)", key, value, par.minimum)
				return
			}
			log.Info().Msgf("MQTT request: Set Parameter %s to %s", key, value)

			// Send the device parameter to Micronova
			writeDevice(par.offset, val, par.mask)
			return
		}
	}
}

func SetUpdateMqtt(update bool) {
	updateMqtt = update
}
