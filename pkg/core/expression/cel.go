package expression

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/thoas/go-funk"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

var (
	stringListType = reflect.TypeOf([]string{})
)

// EvalBooleanExpression evaluates a boolean expression using CEL (e.g. "1 == 1") and returns the result
func EvalBooleanExpression(expression string, context map[string]interface{}) (bool, error) {
	if expression == "" {
		return false, nil
	}

	// init cel go environment
	var exprDecl []*exprpb.Decl
	for key, value := range context {
		switch value.(type) {
		case int:
			exprDecl = append(exprDecl, decls.NewVar(key, decls.Int))
		case string:
			exprDecl = append(exprDecl, decls.NewVar(key, decls.String))
		case []string:
			exprDecl = append(exprDecl, decls.NewVar(key, decls.NewListType(decls.String)))
		case map[string]string:
			exprDecl = append(exprDecl, decls.NewVar(key, decls.NewMapType(decls.String, decls.String)))
		default:
			return false, fmt.Errorf("unsupported evalContext value type: %s", reflect.TypeOf(value))
		}
	}

	celConfig, celConfigErr := cel.NewEnv(
		cel.Declarations(exprDecl...),
		cel.Function("contains",
			cel.Overload("contains_string",
				[]*cel.Type{cel.ListType(cel.StringType), cel.StringType},
				cel.BoolType,
				cel.BinaryBinding(func(lhs, rhs ref.Val) ref.Val {
					list, err := lhs.ConvertToNative(stringListType)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Bool(funk.ContainsString(list.([]string), string(rhs.(types.String))))
				}),
			),
		),
		cel.Function("containsKey",
			cel.Overload("containsKey_map",
				[]*cel.Type{cel.MapType(cel.StringType, cel.StringType), cel.StringType},
				cel.StringType,
				cel.BinaryBinding(func(lhs, rhs ref.Val) ref.Val {
					mapVal, err := lhs.ConvertToNative(reflect.TypeOf(map[string]string{}))
					if err != nil {
						return types.NewErr(err.Error())
					}
					if _, ok := mapVal.(map[string]string)[string(rhs.(types.String))]; ok {
						return types.Bool(true)
					}
					return types.Bool(false)
				}),
			),
		),
		cel.Function("getMapValue",
			cel.Overload("getMapValue_map",
				[]*cel.Type{cel.MapType(cel.StringType, cel.StringType), cel.StringType},
				cel.StringType,
				cel.BinaryBinding(func(lhs, rhs ref.Val) ref.Val {
					mapVal, err := lhs.ConvertToNative(reflect.TypeOf(map[string]string{}))
					if err != nil {
						return types.NewErr(err.Error())
					}
					if value, ok := mapVal.(map[string]string)[string(rhs.(types.String))]; ok {
						return types.String(value)
					} else {
						return types.String("")
					}
				}),
			),
		),
		cel.Function("hasPrefix",
			cel.Overload("hasPrefix_string",
				[]*cel.Type{cel.StringType, cel.StringType},
				cel.BoolType,
				cel.BinaryBinding(func(lhs, rhs ref.Val) ref.Val {
					return types.Bool(strings.HasPrefix(string(lhs.(types.String)), string(rhs.(types.String))))
				}),
			),
		),
	)
	if celConfigErr != nil {
		return false, fmt.Errorf("failed to initialize CEL environment: %w", celConfigErr)
	}

	// prepare program for evaluation
	ast, issues := celConfig.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("failed to compile expression: %w", issues.Err())
	}
	prg, prgErr := celConfig.Program(ast)
	if prgErr != nil {
		return false, fmt.Errorf("failed to construct program: %w", prgErr)
	}

	// evaluate
	execOut, _, execErr := prg.Eval(context)
	if execErr != nil {
		return false, fmt.Errorf("failed to evaluate expression. expr: %s, error: %w", expression, execErr)
	}

	// check result
	if execOut.Type() != types.BoolType {
		return false, fmt.Errorf("expr does not return a boolean")
	}

	return fmt.Sprintf("%+v", execOut) == "true", nil
}