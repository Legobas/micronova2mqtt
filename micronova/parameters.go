package micronova

import (
	"encoding/json"
	"fmt"
	"math"
	"micronova2mqtt/homeassistant"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/sandertv/go-formula/v2"
)

func getParameters() (items []int, masks []int) {
	for _, par := range parameters {
		items = append(items, par.offset)
		masks = append(masks, par.mask)

		log.Trace().Msgf("Parameter %s: item=%v,mask=%v", par.regKey, par.offset, par.mask)
	}
	return
}

func getText(value int, formulaStr string, format string, valueDescr []valueDescr) string {
	var text string
	if len(valueDescr) > 0 {
		for _, descr := range valueDescr {
			if descr.Value == value {
				text = descr.Description
				break
			}
		}
	}

	if text == "" {
		calculatedValue := getCalculated(value, formulaStr)
		formatStr := strings.Replace(format, "{0}", "%v", -1)
		text = fmt.Sprintf(formatStr, calculatedValue)
	}
	return text
}

func getCalculated(value int, formulaStr string) float64 {
	calc := strings.Replace(formulaStr, "#", "x", -1)

	f, err := formula.New(calc)
	if err != nil {
		log.Error().Err(err).Msgf("Formula calculation error: %s", calc)
		return -1
	}
	f.RegisterFunc("int", 1, func(args ...float64) float64 {
		return math.Round(args[0])
	})

	val := formula.Var("x", value)
	return f.MustEval(val)
}

func isActive() bool {
	for _, par := range parameters {
		if par.regKey == "status_get" {
			switch par.text {
			case "Off":
				state = stateInactive
				return false
			case "":
				return false
			default:
				state = stateActive
				return true
			}
		}
	}
	return false
}

func publishParameters() {
	var params map[string]string
	params = make(map[string]string)
	power := "off"

	for _, par := range parameters {
		key := par.regKey
		if par.regKey == "status_get" && par.value != 0 {
			power = "on"
		}
		if len(par.title) != 0 {
			key = par.title
		}
		params[key] = par.text
	}
	jsonBytes, err := json.Marshal(params)
	if err != nil {
		log.Error().Err(err).Msg("Marshal Error")
		return
	}
	jsonStr := string(jsonBytes)
	log.Debug().Msgf("MQTT message: %s", jsonStr)

	topicStart := dm.Config.Mqtt.BaseTopic + "/" + deviceName + "/"
	queue.Publish(topicStart+"parameters", jsonStr)
	// Publish Power state
	queue.Publish(topicStart+"power", power)
}

func publishValues() {
	var values map[string]float64
	values = make(map[string]float64)

	for _, par := range parameters {
		key := par.regKey
		if len(par.title) != 0 {
			key = par.title
		}
		values[key] = par.realValue
	}
	jsonBytes, err := json.Marshal(values)
	if err != nil {
		log.Error().Err(err).Msg("Marshal Error")
		return
	}
	jsonStr := string(jsonBytes)
	log.Debug().Msgf("MQTT message: %s", jsonStr)

	topicStart := dm.Config.Mqtt.BaseTopic + "/" + deviceName + "/"
	queue.Publish(topicStart+"values", jsonStr)
}

func createHAConfig(publisher Queue) homeassistant.HADiscovery {

	var discovery homeassistant.HADiscovery

	discovery.Publisher = publisher

	discovery.DeviceName = deviceName

	device := homeassistant.HADevice{
		Identifiers:  []string{deviceSerial},
		Name:         deviceName,
		Model:        deviceModel,
		Manufacturer: dm.Config.Micronova.Brand,
	}

	for _, par := range parameters {
		var entity homeassistant.HAEntityConfig

		if par.regKey[len(par.regKey)-3:] == "set" {
			if len(par.valueDescr) > 0 {
				entity.Component = "select"
				options := make([]string, 0, len(par.valueDescr)+1)
				for _, value := range par.valueDescr {
					options = append(options, value.Description)
				}

				for i := par.minimum + len(par.valueDescr); i <= par.maximum; i += par.step {
					options = append(options, strconv.Itoa(i))
				}

				entity.Options = options
				entity.StateTopic = fmt.Sprintf("%s/%s/parameters", dm.Config.Mqtt.BaseTopic, deviceName)
			} else {
				entity.Component = "number"
				entity.StateTopic = fmt.Sprintf("%s/%s/values", dm.Config.Mqtt.BaseTopic, deviceName)
			}
			entity.CommandTopic = fmt.Sprintf("%s/%s/command/%s", dm.Config.Mqtt.BaseTopic, deviceName, par.title)
			entity.ValueTemplate = fmt.Sprintf("{{ value_json.%s}}", par.title)
		} else {
			entity.Component = "sensor"
			if len(par.valueDescr) > 0 {
				entity.DeviceClass = "enum"
				options := make([]string, 0, len(par.valueDescr)+1)
				for _, value := range par.valueDescr {
					options = append(options, value.Description)
				}
				for i := par.minimum + len(par.valueDescr); i <= par.maximum; i += par.step {
					options = append(options, strconv.Itoa(i))
				}
				entity.Options = options
				entity.StateTopic = fmt.Sprintf("%s/%s/parameters", dm.Config.Mqtt.BaseTopic, deviceName)
			} else {
				entity.StateTopic = fmt.Sprintf("%s/%s/values", dm.Config.Mqtt.BaseTopic, deviceName)
				if strings.Contains(strings.ToLower(par.regKey), "temp") {
					entity.DeviceClass = "temperature"
				}
			}
			entity.ValueTemplate = fmt.Sprintf("{{ value_json.%s}}", par.title)
		}
		entity.Name = par.title
		entity.ObjectId = par.regKey
		entity.UniqueId = par.regKey
		if entity.Component != "select" {
			entity.Min = 1
			entity.Max = par.maximum
			entity.Step = par.step
		}
		entity.Device = device
		if par.format != "" {
			entity.UnitOfMeasurement = strings.Trim(strings.Replace(par.format, "{0}", "", -1), " ")
		}

		// Select icon
		if strings.Contains(strings.ToLower(par.regKey), "temp") {
			entity.Icon = "mdi:thermometer"
		} else if strings.Contains(strings.ToLower(par.regKey), "calendar") {
			entity.Icon = "mdi:calendar"
		} else if strings.Contains(strings.ToLower(par.regKey), "chrono") {
			entity.Icon = "mdi:calendar-month-outline"
		} else if strings.Contains(strings.ToLower(par.regKey), "clock") {
			entity.Icon = "mdi:clock"
		} else if strings.Contains(strings.ToLower(par.regKey), "alarm") {
			entity.Icon = "mdi:alarm-light"
		} else if strings.Contains(strings.ToLower(par.regKey), "canalization") {
			entity.Icon = "mdi:air-filter"
		} else if strings.Contains(strings.ToLower(par.regKey), "vent") {
			entity.Icon = "mdi:fan"
		} else if strings.Contains(strings.ToLower(par.regKey), "power") {
			entity.Icon = "mdi:fire"
		} else {
			if entity.Component == "number" {
				// set
				entity.Icon = "mdi:cog"
			} else {
				entity.Icon = "mdi:information-variant-circle"
			}
		}

		discovery.Entities = append(discovery.Entities, entity)
	}

	entity := homeassistant.HAEntityConfig{
		Component:    "switch",
		Name:         "Power",
		Icon:         "mdi:power",
		DeviceClass:  "switch",
		UniqueId:     "power",
		CommandTopic: fmt.Sprintf("%s/%s/command/power", dm.Config.Mqtt.BaseTopic, deviceName),
		PayloadOn:    powerOn,
		PayloadOff:   powerOff,
		StateTopic:   fmt.Sprintf("%s/%s/power", dm.Config.Mqtt.BaseTopic, deviceName),
		StateOn:      "on",
		StateOff:     "off",
		Device:       device,
	}

	discovery.Entities = append(discovery.Entities, entity)

	return discovery
}
