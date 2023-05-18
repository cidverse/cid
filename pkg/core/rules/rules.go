package rules

import (
	"strconv"
	"strings"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/expression"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/rs/zerolog/log"
)

const ModuleName = "MODULE_NAME"
const ModuleSlug = "MODULE_SLUG"
const ModuleBuildSystem = "MODULE_BUILD_SYSTEM"
const ModuleBuildSystemSyntax = "MODULE_BUILD_SYSTEM_SYNTAX"
const ModuleFiles = "MODULE_FILES"

// AnyRuleMatches will return true if at least one rule matches, if no rules are provided this always returns true
func AnyRuleMatches(rules []catalog.WorkflowRule, evalContext map[string]interface{}) bool {
	result := 0

	if len(rules) == 0 {
		return true
	}

	for _, rule := range rules {
		if EvaluateRule(rule, evalContext) {
			result++
		}
	}

	return result > 0
}

// EvaluateRulesAsText will check all rules and returns the count of matching rules in the following format: 2/5
func EvaluateRulesAsText(rules []catalog.WorkflowRule, evalContext map[string]interface{}) string {
	matching := EvaluateRules(rules, evalContext)

	return strconv.Itoa(matching) + "/" + strconv.Itoa(len(rules))
}

// EvaluateRules will check all rules and returns the count of matching rules
func EvaluateRules(rules []catalog.WorkflowRule, evalContext map[string]interface{}) int {
	result := 0

	for _, rule := range rules {
		if EvaluateRule(rule, evalContext) {
			result++
		}
	}

	return result
}

// EvaluateRule will evaluate a WorkflowRule and return the result
func EvaluateRule(rule catalog.WorkflowRule, evalContext map[string]interface{}) bool {
	log.Debug().Str("type", string(rule.Type)).Str("expression", rule.Expression).Msg("evaluating rule")

	if rule.Type == "" || rule.Type == catalog.WorkflowExpressionCEL {
		return evalRuleCEL(rule, evalContext)
	}

	log.Error().Str("type", string(rule.Type)).Msg("expression type is not supported!")
	return false
}

func GetRuleContext(env map[string]string) map[string]interface{} {
	return map[string]interface{}{
		"NCI_COMMIT_REF_PATH":        env["NCI_COMMIT_REF_PATH"],
		"NCI_COMMIT_REF_TYPE":        env["NCI_COMMIT_REF_TYPE"],
		"NCI_COMMIT_REF_NAME":        env["NCI_COMMIT_REF_NAME"],
		"NCI_REPOSITORY_HOST_TYPE":   env["NCI_REPOSITORY_HOST_TYPE"],
		"NCI_REPOSITORY_HOST_SERVER": env["NCI_REPOSITORY_HOST_SERVER"],
		"ENV":                        env,
	}
}

func GetModuleRuleContext(env map[string]string, module *analyzerapi.ProjectModule) map[string]interface{} {
	ctx := GetRuleContext(env)

	ctx[ModuleName] = module.Name
	ctx[ModuleSlug] = module.Slug
	ctx[ModuleBuildSystem] = string(module.BuildSystem)
	ctx[ModuleBuildSystemSyntax] = string(module.BuildSystemSyntax)

	var files []string
	for _, file := range module.Files {
		files = append(files, strings.TrimPrefix(strings.TrimPrefix(file, module.Directory+"\\"), module.Directory+"/"))
	}
	ctx[ModuleFiles] = files

	return ctx
}

func evalRuleCEL(rule catalog.WorkflowRule, context map[string]interface{}) bool {
	match, err := expression.EvalBooleanExpression(rule.Expression, context)
	if err != nil {
		return false
	}

	return match
}
