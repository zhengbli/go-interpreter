package evaluator

import (
	"fmt"
	"inter/ast"
	"inter/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func isError(obj object.Object) bool {
	return obj != nil && obj.Type() == object.ERROR_OBJ
}

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch x := node.(type) {
	case *ast.Program:
		return evalProgram(x, env)

	case *ast.PrefixExpression:
		right := Eval(x.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(x.Operator, right)

	case *ast.InfixExpression:
		left := Eval(x.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(x.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(x.Operator, left, right)

	case *ast.ExpressionStatement:
		return Eval(x.Expression, env)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: x.Value}

	case *ast.BooleanLiteral:
		if x.Value {
			return TRUE
		}
		return FALSE

	case *ast.IfExpression:
		cond := Eval(x.Condition, env)
		if isError(cond) {
			return cond
		}
		return evalIfExpression(cond, x.Body, x.ElseBody, env)

	case *ast.BlockStatement:
		return evalBlockStatements(x, env)

	case *ast.ReturnStatement:
		val := Eval(x.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.LetStatement:
		val := Eval(x.Value, env)
		if isError(val) {
			return val
		}
		env.Set(x.Name.Value, val)
		return val

	case *ast.Identifier:
		return evalIdentifier(x, env)

	case *ast.FunctionLiteral:
		params := x.FunctionParameters
		body := x.FunctionBody
		return &object.Function{Parameters: params, Body: body, Env: env}

	case *ast.CallExpression:
		fun := Eval(x.Function, env)
		if isError(fun) {
			return fun
		}

		args := evalExpressions(x.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		funCast := fun.(*object.Function)
		return applyFunction(funCast, args)
	}

	return nil
}

func applyFunction(fun *object.Function, args []object.Object) object.Object {
	newEnv := object.NewEnclosedEnv(fun.Env)
	for paramId, param := range fun.Parameters {
		newEnv.Set(param.Value, args[paramId])
	}

	evaluated := Eval(fun.Body, newEnv)
	if evaluated.Type() == object.RETURNVALUE_OBJ {
		unwrapped := evaluated.(*object.ReturnValue)
		return unwrapped.Value
	}
	return evaluated
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	res := []object.Object{}

	for _, exp := range exps {
		val := Eval(exp, env)
		if isError(val) {
			return []object.Object{val}
		}

		res = append(res, val)
	}
	return res
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: " + node.Value)
	}

	return val
}

func evalIfExpression(cond object.Object, body *ast.BlockStatement, elseBody *ast.BlockStatement, env *object.Environment) object.Object {
	if isTruthy(cond) {
		return Eval(body, env)
	}

	if elseBody != nil {
		return Eval(elseBody, env)
	}

	return NULL
}

func isTruthy(val object.Object) bool {
	switch val {
	case TRUE:
		return true
	case NULL:
		return false
	case FALSE:
		return false
	default:
		return true
	}
}

func evalInfixExpressionForBooleans(operator string, left *object.Boolean, right *object.Boolean) object.Object {
	switch operator {
	case "==":
		return getBool(left.Value == right.Value)
	case "!=":
		return getBool(left.Value != right.Value)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalInfixExpressionForInteger(operator string, left *object.Integer, right *object.Integer) object.Object {
	leftVal := left.Value
	rightVal := right.Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "==":
		return getBool(left.Value == right.Value)
	case "!=":
		return getBool(left.Value != right.Value)
	case ">":
		return getBool(leftVal > rightVal)
	case "<":
		return getBool(leftVal < rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return evalInfixExpressionForInteger(operator, left.(*object.Integer), right.(*object.Integer))
	}

	if left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ {
		return evalInfixExpressionForBooleans(operator, left.(*object.Boolean), right.(*object.Boolean))
	}

	if left.Type() != right.Type() {
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	}

	return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
}

func getBool(val bool) *object.Boolean {
	if val {
		return TRUE
	}
	return FALSE
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperator(right)
	case "-":
		return evalMinusOperator(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalMinusOperator(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	val := right.(*object.Integer).Value
	return &object.Integer{Value: -val}
}

func evalBangOperator(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalBlockStatements(bs *ast.BlockStatement, env *object.Environment) object.Object {
	var res object.Object

	for _, st := range bs.Statements {
		res = Eval(st, env)
		if res.Type() == object.RETURNVALUE_OBJ {
			return res
		}

		if isError(res) {
			return res
		}
	}

	return res
}

func evalProgram(prog *ast.Program, env *object.Environment) object.Object {
	var res object.Object

	for _, st := range prog.Statements {
		res = Eval(st, env)
		if res.Type() == object.RETURNVALUE_OBJ {
			returnVal := res.(*object.ReturnValue)
			return returnVal.Value
		}

		if isError(res) {
			return res
		}
	}

	return res
}

func newError(fmtStr string, args ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(fmtStr, args...)}
}
