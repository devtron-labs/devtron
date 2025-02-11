package standard

func findUnquotedOperators(source string, operator string) int {
	inSingleQuotes := false
	inDoubleQuotes := false
	squareBracketOpen := 0
	squareBracketClose := 0
	roundBracketOpen := 0
	roundBracketClose := 0

	inSquareBrackets := func() bool {
		return squareBracketOpen > squareBracketClose
	}
	inRoundBrackets := func() bool {
		return roundBracketOpen > roundBracketClose
	}

	for idx, rne := range source {
		switch rne {
		case '(':
			if inSingleQuotes || inDoubleQuotes || inSquareBrackets() {
				continue
			}
			roundBracketOpen++
			break
		case ')':
			if inSingleQuotes || inDoubleQuotes || inSquareBrackets() {
				continue
			}
			roundBracketClose++
			break
		case '[':
			if inSingleQuotes || inDoubleQuotes || inRoundBrackets() {
				continue
			}
			squareBracketOpen++
			break
		case ']':
			if inSingleQuotes || inDoubleQuotes || inRoundBrackets() {
				continue
			}
			squareBracketClose++
			break
		case '\'':
			if inDoubleQuotes || inSquareBrackets() || inRoundBrackets() {
				continue
			}
			inSingleQuotes = !inSingleQuotes
			break
		case '"':
			if inSingleQuotes || inSquareBrackets() || inRoundBrackets() {
				continue
			}
			inDoubleQuotes = !inDoubleQuotes
			break
		default:
			break
		}

		if inSingleQuotes || inDoubleQuotes || inSquareBrackets() || inRoundBrackets() {
			continue
		}

		if len(operator) > idx+1 {
			// do not try and get a two character
			// operator on first idx
			continue
		}

		start := (idx + 1) - (len(operator))
		end := idx
		token := source[start : end+1]
		if token == operator {
			return start
		}
	}

	return -1
}
