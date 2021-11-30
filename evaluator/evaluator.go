package evaluator

import (
	"inter/ast"
	"inter/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node) object.Object {
	switch x := node.(type) {
	case *ast.Program:
		return evalStatements(x.Statements)
	case *ast.PrefixExpression:
		right := Eval(x.Right)
		return evalPrefixExpression(x.Operator, right)
	case *ast.InfixExpression:
		left := Eval(x.Left)
		right := Eval(x.Right)
		return evalInfixExpression(x.Operator, left, right)
	case *ast.ExpressionStatement:
		return Eval(x.Expression)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: x.Value}
	case *ast.BooleanLiteral:
		if x.Value {
			return TRUE
		}
		return FALSE
	case *ast.IfExpression:
		cond := Eval(x.Condition)
		return evalIfExpression(cond, x.Body, x.ElseBody)
	case *ast.BlockStatement:
		return evalStatements(x.Statements)
	}

	return nil
}

func evalIfExpression(cond object.Object, body *ast.BlockStatement, elseBody *ast.BlockStatement) object.Object {
	if isTruthy(cond) {
		return Eval(body)
	}

	if elseBody != nil {
		return Eval(elseBody)
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
		return NULL
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
	case "\\":
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
		return NULL
	}
}

func evalInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return evalInfixExpressionForInteger(operator, left.(*object.Integer), right.(*object.Integer))
	}

	if left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ {
		return evalInfixExpressionForBooleans(operator, left.(*object.Boolean), right.(*object.Boolean))
	}

	return NULL
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
		return NULL
	}
}

func evalMinusOperator(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return NULL
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

func evalStatements(sts []ast.Statement) object.Object {
	var res object.Object

	for _, st := range sts {
		res = Eval(st)
	}

	return res
}
