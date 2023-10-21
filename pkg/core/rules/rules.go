package rules

import (
	"github.com/PhilippHeuer/tmux-tms/pkg/core/expression"
	"github.com/rs/zerolog/log"
)

// EvaluateRules will check all rules and returns the count of matching rules
func EvaluateRules(rules []string, evalContext map[string]interface{}) int {
	result := 0

	for _, rule := range rules {
		if EvaluateRule(rule, evalContext) {
			result++
		}
	}

	return result
}

// EvaluateRule will evaluate a WorkflowRule and return the result
func EvaluateRule(rule string, evalContext map[string]interface{}) bool {
	log.Debug().Str("expression", rule).Msg("evaluating rule")

	match, err := expression.EvalBooleanExpression(rule, evalContext)
	if err != nil {
		log.Debug().Err(err).Str("expression", rule).Msg("failed to evaluate expression")
		return false
	}

	return match
}
