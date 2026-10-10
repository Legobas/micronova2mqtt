package micronova

import (
	"fmt"
	"micronova2mqtt/files"
	"slices"

	"github.com/rs/zerolog/log"
)

var activated []string

type ActionsHolder struct {
	actions []files.Action
}

func (auto *ActionsHolder) reset() {
	activated = []string{}
}

func (auto *ActionsHolder) Process(actions []files.Action) {
	auto.actions = actions

	for _, action := range actions {
		triggerKey := action.Trigger.GetKey
		triggerMinValue := action.Trigger.MinimumValue
		triggerMaxValue := action.Trigger.MaximumValue
		triggerAct := fmt.Sprintf("%s is between %d and %d", triggerKey, triggerMinValue, triggerMaxValue)
		// check if action is already activated
		if !slices.Contains(activated, triggerAct) {
			// get real value
			var realValue = 0
			for _, par := range parameters {
				if par.regKey == triggerKey {
					realValue = par.value
					break
				}
			}

			log.Debug().Msgf("Action: %s, real value: %d", triggerAct, realValue)

			// set action values
			if realValue >= triggerMinValue && realValue <= triggerMaxValue {
				for _, setValue := range action.SetValues {
					SetParameterByRegKey(setValue.SetKey, setValue.Value)
				}
				log.Info().Msgf("Action activated: %s", triggerAct)
				activated = append(activated, triggerAct)
				break
			}
		}
	}
}
