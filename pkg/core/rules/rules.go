package rules

import (
	"fmt"
	"strconv"

	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/cidverse/normalizeci/pkg/ncispec"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/rs/zerolog/log"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

const ModuleName = "MODULE_NAME"
const ModuleSlug = "MODULE_SLUG"
const ModuleBuildSystem = "MODULE_BUILD_SYSTEM"
const ModuleBuildSystemSyntax = "MODULE_BUILD_SYSTEM_SYNTAX"

// AnyRuleMatches will return true if at least one rule matches, if no rules are provided this always returns true
func AnyRuleMatches(rules []config.WorkflowRule, evalContext map[string]interface{}) bool {
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
func EvaluateRulesAsText(rules []config.WorkflowRule, evalContext map[string]interface{}) string {
	matching := EvaluateRules(rules, evalContext)

	return strconv.Itoa(matching) + "/" + strconv.Itoa(len(rules))
}

// EvaluateRules will check all rules and returns the count of matching rules
func EvaluateRules(rules []config.WorkflowRule, evalContext map[string]interface{}) int {
	result := 0

	for _, rule := range rules {
		if EvaluateRule(rule, evalContext) {
			result++
		}
	}

	return result
}

// EvaluateRule will evaluate a WorkflowRule and return the result
func EvaluateRule(rule config.WorkflowRule, evalContext map[string]interface{}) bool {
	log.Debug().Str("type", string(rule.Type)).Str("expression", rule.Expression).Msg("evaluating rule")

	if rule.Type == "" || rule.Type == config.WorkflowExpressionCEL {
		return evalRuleCEL(rule, evalContext)
	}

	log.Error().Str("type", string(rule.Type)).Msg("expression type is not supported!")
	return false
}

func GetRuleContext(env map[string]string) map[string]interface{} {
	return map[string]interface{}{
		ncispec.NCI_COMMIT_REF_PATH: env[ncispec.NCI_COMMIT_REF_PATH],
		ncispec.NCI_COMMIT_REF_TYPE: env[ncispec.NCI_COMMIT_REF_TYPE],
		ncispec.NCI_COMMIT_REF_NAME: env[ncispec.NCI_COMMIT_REF_NAME],
	}
}

func GetModuleRuleContext(env map[string]string, module *analyzerapi.ProjectModule) map[string]interface{} {
	ctx := GetRuleContext(env)

	ctx[ModuleName] = module.Name
	ctx[ModuleSlug] = module.Slug
	ctx[ModuleBuildSystem] = string(module.BuildSystem)
	ctx[ModuleBuildSystemSyntax] = string(module.BuildSystemSyntax)

	return ctx
}

func evalRuleCEL(rule config.WorkflowRule, evalContext map[string]interface{}) bool {
	if rule.Expression == "" {
		return false
	}

	// init cel go environment
	var exprDecl []*exprpb.Decl
	for key, value := range evalContext {
		switch value.(type) {
		case int:
			exprDecl = append(exprDecl, decls.NewVar(key, decls.Int))
		case string:
			exprDecl = append(exprDecl, decls.NewVar(key, decls.String))
		default:
			log.Fatal().Str("type", string(rule.Type)).Str("key", key).Interface("value", value).Msg("unsupported evalContext value type")
		}
	}
	celConfig, celConfigErr := cel.NewEnv(cel.Declarations(exprDecl...))
	if celConfigErr != nil {
		log.Fatal().Err(celConfigErr).Msg("failed to initialize CEL environment")
	}

	// prepare program for evaluation
	ast, issues := celConfig.Compile(rule.Expression)
	if issues != nil && issues.Err() != nil {
		log.Fatal().Err(issues.Err()).Msg("stage rule type error: " + issues.Err().Error())
	}
	prg, prgErr := celConfig.Program(ast)
	if prgErr != nil {
		log.Fatal().Err(prgErr).Msg("program construction error")
	}

	// evaluate
	execOut, _, execErr := prg.Eval(evalContext)
	if execErr != nil {
		log.Warn().Err(execErr).Msg("failed to evaluate filter rule")
	}

	// check result
	if execOut.Type() != types.BoolType {
		log.Error().Str("type", string(rule.Type)).Str("expression", rule.Expression).Msg("rule expr does not return a boolean")
		return false
	}

	return fmt.Sprintf("%+v", execOut) == "true"
}
