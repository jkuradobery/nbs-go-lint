package example

import "fmt"

// This is a file to check the linter for line break after
// multiple line function signature
// expected analysis results for test are set via want comments


// This is valid
//goland:noinspection GoUnusedFunction
func functionExample(a, b int) int {
	return a + b
}

// This is valid
//goland:noinspection GoUnusedFunction
func someFunctionWithLongYetValidSignature(
	firstUnnecessaryLongArgumentName int,
	secondUnnecessaryLongArgumentName int) {

	fmt.Printf(
		"OutputsSomeArguments %d: %d",
		firstUnnecessaryLongArgumentName,
		secondUnnecessaryLongArgumentName,
	)
}

// This is also valid
//goland:noinspection GoUnusedFunction
func someFunctionWithLongYetValidSignatureWithComment(
	firstUnnecessaryLongArgumentNameWithComment int,
	secondUnnecessaryLongArgumentNameWithComment int) {

	// some comment to show that I know what I am doing
	fmt.Printf(
		"OutputsSomeArguments (But we have acomment) %d: %d",
		firstUnnecessaryLongArgumentNameWithComment,
		secondUnnecessaryLongArgumentNameWithComment,
	)
}

// This is invalid
// WARNING: This type of formatting is forbidden by gofmt, alas we still want to check this.
// DO NOT FORMAT THIS FILE WITH GOFMT
// @formatter:off
//goland:noinspection GoUnusedFunction
func someFunctionWithLongYetInvalidSignature(
	firstUnnecessaryLongArgumentName int,
	secondUnnecessaryLongArgumentName int) { //want `Too many line breaks after the multiline function signature: 2`


	fmt.Print(
		"Hello world",
		firstUnnecessaryLongArgumentName+secondUnnecessaryLongArgumentName,
	)
}
// @formatter:on

// This is invalid
//goland:noinspection GoUnusedFunction
func someFunctionWithStatementAfterClosingBracket(
	firstUnnecessaryLongArgumentName int,
	secondUnnecessaryLongArgumentName int) { //want `Line break after multiline function signature is required`
	fmt.Println("Hello world")
	fmt.Print(
		firstUnnecessaryLongArgumentName,
		secondUnnecessaryLongArgumentName)
}

// This is valid
//goland:noinspection GoUnusedFunction,GoUnusedParameter
func someFunctionWithEmptyBody(
	firstUselessArgument []string,
	secondUselessArgument map[string]string,
){

}

// This is invalid
//goland:noinspection GoUnusedFunction,GoUnusedParameter
func someFunctionWithEmptyBodyFailsLinter(
	firstUselessArgument []string,
	secondUselessArgument map[string]string,
){} // want `Line break after multiline function signature is required`

// This is invalid
// @formatter:off
//goland:noinspection GoUnusedFunction,GoUnusedParameter
func someFunctionWithEmptyBodyMultipleLineInside(
	firstUselessArgument []string,
	secondUselessArgument map[string]string,
){ //want `Too many line breaks after the multiline function signature: 3`



}
// @formatter:on
