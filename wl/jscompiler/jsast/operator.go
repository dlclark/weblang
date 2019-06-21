package jsast

// Operator order from Javascript
const (
	LowestPrec = 0 // non-operators

	UnaryPrec   = 16
	HighestPrec = 21
)

// Precedence returns the operator precedence of the binary
// operator op. If op is not a binary operator, the result
// is LowestPrecedence.
//
// This is not an exhaustive view of Javascript operators
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Operator_Precedence
func Precedence(op string) int {
	switch op {
	case "**":
		return 15
	case "*", "/", "%":
		return 14
	case "+", "-":
		return 13
	case "<<", ">>", ">>>":
		return 12
	case "<", "<=", ">", ">=":
		return 11
	case "==", "!=", "===", "!==":
		return 10
	case "&":
		return 9
	case "^":
		return 8
	case "|":
		return 7
	case "&&":
		return 6
	case "||":
		return 5
	}
	return LowestPrec
}
