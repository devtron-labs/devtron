package govaluate

import (
	"errors"
	"fmt"
)

type lexerState struct {
	isEOF          bool
	isNullable     bool
	kind           TokenKind
	validNextKinds []TokenKind
}

// lexer states.
// Constant for all purposes except compiler.
var validLexerStates = []lexerState{

	{
		kind:       UNKNOWN,
		isEOF:      false,
		isNullable: true,
		validNextKinds: []TokenKind{

			PREFIX,
			NUMERIC,
			BOOLEAN,
			VARIABLE,
			PATTERN,
			FUNCTION,
			ACCESSOR,
			STRING,
			TIME,
			CLAUSE,
		},
	},

	{

		kind:       CLAUSE,
		isEOF:      false,
		isNullable: true,
		validNextKinds: []TokenKind{

			PREFIX,
			NUMERIC,
			BOOLEAN,
			VARIABLE,
			PATTERN,
			FUNCTION,
			ACCESSOR,
			STRING,
			TIME,
			CLAUSE,
			CLAUSE_CLOSE,
		},
	},

	{

		kind:       CLAUSE_CLOSE,
		isEOF:      true,
		isNullable: true,
		validNextKinds: []TokenKind{

			COMPARATOR,
			MODIFIER,
			NUMERIC,
			BOOLEAN,
			VARIABLE,
			STRING,
			PATTERN,
			TIME,
			CLAUSE,
			CLAUSE_CLOSE,
			LOGICALOP,
			TERNARY,
			SEPARATOR,
		},
	},

	{

		kind:       NUMERIC,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []TokenKind{

			MODIFIER,
			COMPARATOR,
			LOGICALOP,
			CLAUSE_CLOSE,
			TERNARY,
			SEPARATOR,
		},
	},
	{

		kind:       BOOLEAN,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []TokenKind{

			MODIFIER,
			COMPARATOR,
			LOGICALOP,
			CLAUSE_CLOSE,
			TERNARY,
			SEPARATOR,
		},
	},
	{

		kind:       STRING,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []TokenKind{

			MODIFIER,
			COMPARATOR,
			LOGICALOP,
			CLAUSE_CLOSE,
			TERNARY,
			SEPARATOR,
		},
	},
	{

		kind:       TIME,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []TokenKind{

			MODIFIER,
			COMPARATOR,
			LOGICALOP,
			CLAUSE_CLOSE,
			SEPARATOR,
		},
	},
	{

		kind:       PATTERN,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []TokenKind{

			MODIFIER,
			COMPARATOR,
			LOGICALOP,
			CLAUSE_CLOSE,
			SEPARATOR,
		},
	},
	{

		kind:       VARIABLE,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []TokenKind{

			MODIFIER,
			COMPARATOR,
			LOGICALOP,
			CLAUSE_CLOSE,
			TERNARY,
			SEPARATOR,
		},
	},
	{

		kind:       MODIFIER,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []TokenKind{

			PREFIX,
			NUMERIC,
			VARIABLE,
			FUNCTION,
			ACCESSOR,
			STRING,
			BOOLEAN,
			CLAUSE,
			CLAUSE_CLOSE,
		},
	},
	{

		kind:       COMPARATOR,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []TokenKind{

			PREFIX,
			NUMERIC,
			BOOLEAN,
			VARIABLE,
			FUNCTION,
			ACCESSOR,
			STRING,
			TIME,
			CLAUSE,
			CLAUSE_CLOSE,
			PATTERN,
		},
	},
	{

		kind:       LOGICALOP,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []TokenKind{

			PREFIX,
			NUMERIC,
			BOOLEAN,
			VARIABLE,
			FUNCTION,
			ACCESSOR,
			STRING,
			TIME,
			CLAUSE,
			CLAUSE_CLOSE,
		},
	},
	{

		kind:       PREFIX,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []TokenKind{

			NUMERIC,
			BOOLEAN,
			VARIABLE,
			FUNCTION,
			ACCESSOR,
			CLAUSE,
			CLAUSE_CLOSE,
		},
	},

	{

		kind:       TERNARY,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []TokenKind{

			PREFIX,
			NUMERIC,
			BOOLEAN,
			STRING,
			TIME,
			VARIABLE,
			FUNCTION,
			ACCESSOR,
			CLAUSE,
			SEPARATOR,
		},
	},
	{

		kind:       FUNCTION,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []TokenKind{
			CLAUSE,
		},
	},
	{

		kind:       ACCESSOR,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []TokenKind{
			CLAUSE,
			MODIFIER,
			COMPARATOR,
			LOGICALOP,
			CLAUSE_CLOSE,
			TERNARY,
			SEPARATOR,
		},
	},
	{

		kind:       SEPARATOR,
		isEOF:      false,
		isNullable: true,
		validNextKinds: []TokenKind{

			PREFIX,
			NUMERIC,
			BOOLEAN,
			STRING,
			TIME,
			VARIABLE,
			FUNCTION,
			ACCESSOR,
			CLAUSE,
		},
	},
}

func (l lexerState) canTransitionTo(kind TokenKind) bool {

	for _, validKind := range l.validNextKinds {

		if validKind == kind {
			return true
		}
	}

	return false
}

func checkExpressionSyntax(tokens []ExpressionToken) error {

	var state lexerState
	var lastToken ExpressionToken
	var err error

	state = validLexerStates[0]

	for _, token := range tokens {

		if !state.canTransitionTo(token.Kind) {

			// call out a specific error for tokens looking like they want to be functions.
			if lastToken.Kind == VARIABLE && token.Kind == CLAUSE {
				return errors.New("Undefined function " + lastToken.Value.(string))
			}

			firstStateName := fmt.Sprintf("%s [%v]", state.kind.String(), lastToken.Value)
			nextStateName := fmt.Sprintf("%s [%v]", token.Kind.String(), token.Value)

			return errors.New("Cannot transition token types from " + firstStateName + " to " + nextStateName)
		}

		state, err = getLexerStateForToken(token.Kind)
		if err != nil {
			return err
		}

		if !state.isNullable && token.Value == nil {

			errorMsg := fmt.Sprintf("token kind '%v' cannot have a nil value", token.Kind.String())
			return errors.New(errorMsg)
		}

		lastToken = token
	}

	if !state.isEOF {
		return errors.New("unexpected end of expression")
	}
	return nil
}

func getLexerStateForToken(kind TokenKind) (lexerState, error) {

	for _, possibleState := range validLexerStates {

		if possibleState.kind == kind {
			return possibleState, nil
		}
	}

	errorMsg := fmt.Sprintf("no lexer state found for token kind '%v'\n", kind.String())
	return validLexerStates[0], errors.New(errorMsg)
}
