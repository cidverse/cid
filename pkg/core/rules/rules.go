package rules

import (
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/go-rules/pkg/expr"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/rs/zerolog/log"
)

const (
	ModuleName              = "MODULE_NAME"
	ModuleSlug              = "MODULE_SLUG"
	ModuleType              = "MODULE_TYPE"
	ModuleBuildSystem       = "MODULE_BUILD_SYSTEM"
	ModuleBuildSystemSyntax = "MODULE_BUILD_SYSTEM_SYNTAX"
	ModuleConfigType        = "MODULE_CONFIG_TYPE"
	ModuleSpecificationType = "MODULE_SPECIFICATION_TYPE"
	ModuleDeploymentSpec    = "MODULE_DEPLOYMENT_SPEC"
	ModuleDeploymentType    = "MODULE_DEPLOYMENT_TYPE"
	ModuleFiles             = "MODULE_FILES"
)

// AnyRuleMatches will return true if at least one rule matches, if no rules are provided this always returns true
func AnyRuleMatches(rules []catalog.WorkflowRule, evalContext map[string]interface{}) bool {
	if len(rules) == 0 {
		return true
	}

	for _, rule := range rules {
		if EvaluateRule(rule, evalContext) {
			return true
		}
	}

	return false
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
	slog.With("type", string(rule.Type), "expression", rule.Expression).Debug("evaluating rule")

	if rule.Type == "" || rule.Type == catalog.WorkflowExpressionCEL {
		return evalRuleCEL(rule, evalContext)
	}

	slog.With("type", string(rule.Type)).Error("expression type is not supported!")
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

func GetProjectRuleContext(env map[string]string, modules []*analyzerapi.ProjectModule) map[string]interface{} {
	var buildSystems []string
	var specificationTypes []string
	var configTypes []string
	var deploymentTypes []string
	var languages []string
	for _, module := range modules {
		if string(module.BuildSystem) != "" && !slices.Contains(buildSystems, string(module.BuildSystem)) {
			buildSystems = append(buildSystems, string(module.BuildSystem))
		}
		if string(module.SpecificationType) != "" && !slices.Contains(specificationTypes, string(module.SpecificationType)) {
			specificationTypes = append(specificationTypes, string(module.SpecificationType))
		}
		if string(module.ConfigType) != "" && !slices.Contains(configTypes, string(module.ConfigType)) {
			configTypes = append(configTypes, string(module.ConfigType))
		}
		if string(module.DeploymentType) != "" && !slices.Contains(deploymentTypes, string(module.DeploymentType)) {
			deploymentTypes = append(deploymentTypes, string(module.DeploymentType))
		}

		for _, lang := range module.Language {
			if !slices.Contains(languages, lang) {
				languages = append(languages, lang)
			}
		}
	}

	rc := GetRuleContext(env)
	rc["PROJECT_BUILD_SYSTEMS"] = buildSystems
	rc["PROJECT_SPECIFICATION_TYPES"] = specificationTypes
	rc["PROJECT_CONFIG_TYPES"] = configTypes
	rc["PROJECT_DEPLOYMENT_TYPES"] = deploymentTypes
	rc["PROJECT_LANGUAGES"] = languages

	return rc
}

func GetModuleRuleContext(env map[string]string, module *analyzerapi.ProjectModule) map[string]interface{} {
	ctx := GetRuleContext(env)
	ctx[ModuleName] = module.Name
	ctx[ModuleSlug] = module.Slug
	ctx[ModuleType] = string(module.Type)
	ctx[ModuleBuildSystem] = string(module.BuildSystem)
	ctx[ModuleBuildSystemSyntax] = string(module.BuildSystemSyntax)
	ctx[ModuleConfigType] = string(module.ConfigType)
	ctx[ModuleSpecificationType] = string(module.SpecificationType)
	ctx[ModuleDeploymentSpec] = string(module.DeploymentSpec)
	ctx[ModuleDeploymentType] = module.DeploymentType

	var files []string
	for _, file := range module.Files {
		files = append(files, strings.TrimPrefix(strings.TrimPrefix(file, module.Directory+"\\"), module.Directory+"/"))
	}
	ctx[ModuleFiles] = files

	return ctx
}

func evalRuleCEL(rule catalog.WorkflowRule, context map[string]interface{}) bool {
	match, err := expr.EvalBooleanExpression(rule.Expression, context)
	if err != nil {
		log.Debug().Err(err).Str("expression", rule.Expression).Msg("failed to evaluate workflow rule expression")
		return false
	}

	return match
}
