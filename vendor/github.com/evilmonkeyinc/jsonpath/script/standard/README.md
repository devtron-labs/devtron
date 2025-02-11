# Standard Script Engine

The standard script engine is a basic implementation of the script.Engine interface that is used as the default script engine

## Supported Operations

|operator|name|supported types|description|
|-|-|-|-|
|`\|\|`|logical OR|boolean|return true if left-side OR right-side are true|
|`&&`|logical AND|boolean|return true if left-side AND right-side are true|
|`!`|not|boolean|return true if right-side is false. The is no left-side argument|
|`==`|equals|number\|string|return true if left-side and right-side arguments are equal|
|`!=`|not equals|number\|string|return true if left-side and right-side arguments are not equal|
|`<=`|less than or equal to|number|return true if left-side number is less than or equal to the right-side number|
|`>=`|greater than or equal to|number|return true if left-side number is greater than or equal to the right-side number|
|`<`|less than|number|return true if left-side number is less than the right-side number|
|`>`|greater than|number|return true if left-side number is greater than the right-side number|
|`=~`|regex|string|perform a regex match on the left-side value using the right-side pattern|
|`+`|plus/addition|number|return the left-side number added to the right-side number|
|`-`|minus/subtraction|number|return the left-side number minus the right-side number|
|`**`|power|number|return the left-side number increased to the power of the right-side number|
|`*`|multiplication|number|return the left-side number multiplied by the right-side number|
|`/`|division|number|return the left-side number divided by the right-side number|
|`%`|modulus|integer|return the remainder of the left-side number divided by the right-side number|
|`in`|in|number\|string\|boolean and collection|return true if the left-side argument is in the right-side collection|
|`not in`|in|number\|string\|boolean and collection|return true if the left-side argument is not in the right-side collection|

All operators have a left-side and right-side argument, expect the not `!` operator which only as a right-side argument. The arguments can be strings, numbers, boolean values, arrays, objects, a special parameter, or other expressions, for example `true && true || false` includes the logical AND operator with left-side `true` and right-side a logical OR operator with left-side `true` and right-side `false`.

### Regex

The regex operator will perform a regex match check using the left side argument as the input and the right as the regex pattern.

The right side pattern should be passed as a string, between single or double quotes, to ensure that no characters are mistaken for other operators.

> the regex operation is handled by the standard [`regexp`](https://pkg.go.dev/regexp) golang library `Match` function.

### In and Not In

The `in` and `not in` operators will check if the left-side value, treated either as a number, string, or boolean value, is included in the right-side collection values, a collection can either be an array, slice, or the values of a map.

## Special Parameters

The following symbols/tokens have special meaning when used in script expressions and will be replaced before the expression is evaluated. The symbols used within a string, between single or double quotes, will not be replaced.

|symbol|name|replacement|
|-|-|-|
|`$`|root|the root json data node|
|`@`|current|the current json data node|
|`nil`|nil|`nil`|
|`null`|null|`nil`|

Using the root or current symbol allows to embed a JSONPath selector within an expression and it is expected that any argument that includes these characters should be a valid selector.

The nil and null tokens can be used interchangeably to represent a nil value.

> remember that the @ character has different meaning in subscripts than it does in filters.

## Limitations

The script parser does not infer meaning from symbols/tokens and the neighboring characters, what may be considered a valid mathematical equation is not always a valid script expression.

For example, the equations `8+3-3*6+2(2*3)` will not be parsed correctly as the engine does not under stand that `2(2*3)` is the same as `2*(2*3)`, if you needed such an equation you would remove any ambiguity of what a number next to a bracket means like so `8+3-3*6+2*(2*3)`.
