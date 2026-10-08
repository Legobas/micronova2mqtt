package micronova

import (
	"micronova2mqtt/files"
	"slices"
	"strconv"

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
	log.Debug().Msgf("Actions: %v", auto.actions)

	for _, action := range actions {
		triggerKey := action.Trigger.GetKey
		triggerValue := action.Trigger.MinimumValue
		triggerAct := triggerKey + "=" + strconv.Itoa(triggerValue)
		// check if action is already activated
		if slices.Contains(activated, triggerAct) {
			continue
		}

		// get real value
		var realValue = 0
		for _, par := range parameters {
			if par.regKey == triggerKey {
				realValue = par.value
				break
			}
		}

		// set action values
		if realValue >= triggerValue {
			for _, setValue := range action.SetValues {
				SetParameterByRegKey(setValue.SetKey, setValue.Value)
			}
			log.Info().Msgf("Action activated: %s", triggerAct)
			activated = append(activated, triggerAct)
		}
	}
}
