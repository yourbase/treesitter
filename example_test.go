package sitter_test

import (
	"fmt"

	sitter "github.com/yourbase/treesitter"
	"github.com/yourbase/treesitter/json"
)

// Simple program that uses the JSON parser.
// Equivalent to the first tree-sitter example: https://tree-sitter.github.io/tree-sitter/using-parsers#an-example-program
func Example() {
	// Create a parser.
	parser := sitter.NewParser()
	defer parser.Close()

	// Set the parser's language (JSON in this case).
	parser.SetLanguage(json.GetLanguage())

	// Build a syntax tree based on source code stored in a string.
	const sourceCode = "[1, null]"
	tree := parser.Parse(nil, []byte(sourceCode))
	defer tree.Close()

	// Get the root node of the syntax tree.
	rootNode := tree.RootNode()

	// Get some child nodes.
	arrayNode := rootNode.NamedChild(0)
	numberNode := arrayNode.NamedChild(0)

	// Check that the nodes have the expected types and child counts.
	fmt.Println(rootNode.Type(), rootNode.ChildCount())
	fmt.Println(arrayNode.Type(), arrayNode.ChildCount(), arrayNode.NamedChildCount())
	fmt.Println(numberNode.Type(), numberNode.ChildCount())

	// Print the syntax tree as an S-expression.
	fmt.Println("Syntax tree:", rootNode)

	// Output:
	// document 1
	// array 5 2
	// number 0
	// Syntax tree: (document (array (number) (null)))
}
