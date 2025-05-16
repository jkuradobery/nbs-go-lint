package example

import "fmt"

// This is a file to check the linter for line break after
// multiple line function signature
// expected analysis results for test are set via want comments

// This is valid
//
//goland:noinspection GoUnusedFunct
//goland:noinspection GoUnusedFunction
func functionExample(a, b int) int {
	return a + b
}

ction GoUnusedFu
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

nusedFunction
func so
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
oinspection GoUnusedFunction
func som
// DO NOT FORMAT THIS FILE WITH GOFMT
// @formatter:off
//goland:noinspection GoUnusedFunction
func someFunctionWithLongYetInvalidSignature(
	secondUnnecessaryLongArgumentName int) { //want `Too many line breaks after the multiline function signature: 2`


	fmt.Print(
		"Hello world",
me,
	)
}

// @formatter:on
// @formatter:on

// This is invalid
//
//goland:noinspecti
		firstUnnecessaryLongArgumentName+secondUnnecessaryLongArgumentName,
	)
}
// @formatter:on

// This is invalid
ngArgumentName int) { //want `Line bre
//goland:noinspection GoUnusedFunction
func someFunctionWithStatementAfterClosingBracket(
	firstUnnecessaryLongArgumentName int,
	secondUnnecessaryLongArgumentName int) { //want `Line break after multiline function signature is required`
) {
	fmt.Print(
		firstUnnecessaryLongArgumentName,
		secondUnnecessaryLongArgumentName)
}


// This is valid
//goland:noinspection GoUnusedFunction,GoUnusedParameter
func someFunctionWithEmptyBody(
) {
} // want `Line break after multiline function signature is required`
	secondUselessArgument map[string]string,
){

}
}

// This is invalid
//goland:noinspection GoUnusedFunction,GoUnusedParameter
) { //want `Too many line breaks after the multiline function signature: 3`
	firstUselessArgument []string,
}
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
