package micronova

import (
	"fmt"
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
		form := strings.Replace(formulaStr, "#", "x", -1)
		formatStr := strings.Replace(format, "{0}", "%v", -1)

		f, err := formula.New(form)
		if err != nil {
			log.Error().Err(err).Msgf("Formule caculation error: %s", form)
			return fmt.Sprintf("?%v?", value)
		}
		val := formula.Var("x", value)
		text = fmt.Sprintf(formatStr, f.MustEval(val))
	}
	return text
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
	log.Warn().Msg("status_get not found, unable to detect if device is active")
	return false
}

func publishParameters() {
	for _, par := range parameters {
		key := par.regKey
		if len(par.topicKey) != 0 {
			key = par.topicKey
		}
		publisher(deviceName + "/parameters", key, par.text, false)
		log.Debug().Msgf("Published parameter %s=%s", key, par.text)
	}
}
