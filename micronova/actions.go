package micronova

import (
	"fmt"
	"micronova2mqtt/files"
	"slices"

	"github.com/rs/zerolog/log"
)

type ActionsHolder struct {
	actions              []files.Action
	activated            []string
	getCalculated        func(value int, formulaStr string) float64
	setParameterByRegKey func(key string, value int)
	parameters           *[]parameter
}

func (holder *ActionsHolder) reset() {
	holder.activated = []string{}
}

func (holder *ActionsHolder) Process(actions []files.Action) {
	holder.actions = actions

	for _, action := range actions {
		triggerKey := action.Trigger.GetKey
		triggerMinValue := action.Trigger.MinimumValue
		triggerMaxValue := action.Trigger.MaximumValue
		triggerAct := fmt.Sprintf("%s >= %.1f & <= %.1f", triggerKey, triggerMinValue, triggerMaxValue)
		// check if action is already activated
		if !slices.Contains(holder.activated, triggerAct) {
			// get real value
			var realValue float64
			for _, par := range *holder.parameters {
				if par.regKey == triggerKey {
					realValue = par.realValue
					break
				}
			}

			log.Debug().Msgf("Action: %s, real value: %.1f", triggerAct, realValue)

			// set action values
			if realValue >= triggerMinValue && realValue <= triggerMaxValue {
				for _, setValue := range action.SetValues {
					holder.setParameterByRegKey(setValue.SetKey, setValue.Value)
				}
				log.Info().Msgf("Action activated: %s", triggerAct)
				holder.activated = append(holder.activated, triggerAct)
				break
			}
		}
	}
}
