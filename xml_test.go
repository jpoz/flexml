package flexml

import (
	"strings"
	"testing"
)

func TestParseValidXML(t *testing.T) {
	// Test the basic case from the example in the requirements
	xml := `<response><message>Greetings</message></response>`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	nodes, ok := doc.DeepFind("message")
	if !ok {
		t.Fatal("Failed to find message element")
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 message node, got %d", len(nodes))
	}

	if nodes[0].GetText() != "Greetings" {
		t.Fatalf("Expected text 'Greetings', got '%s'", nodes[0].GetText())
	}
}

func TestParsePartialXML(t *testing.T) {
	// Test with unclosed tag as in the requirements
	xml := `<key>Hello`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	nodes, ok := doc.DeepFind("key")
	if !ok {
		t.Fatal("Failed to find key element")
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 key node, got %d", len(nodes))
	}

	if nodes[0].GetText() != "Hello" {
		t.Fatalf("Expected text 'Hello', got '%s'", nodes[0].GetText())
	}
}

func TestParseInvalidXML(t *testing.T) {
	// Test with text before XML as in the requirements
	xml := `Hello how are you
<name>James</name>`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Check that the text is preserved
	if doc.Root.Children[0].Type != TextNode {
		t.Fatal("First child should be a text node")
	}

	if doc.Root.Children[0].Value != "Hello how are you\n" {
		t.Fatalf("Expected text 'Hello how are you\\n', got '%s'", doc.Root.Children[0].Value)
	}

	// Check that the name element is found
	nodes, ok := doc.DeepFind("name")
	if !ok {
		t.Fatal("Failed to find name element")
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 name node, got %d", len(nodes))
	}

	if nodes[0].GetText() != "James" {
		t.Fatalf("Expected text 'James', got '%s'", nodes[0].GetText())
	}
}

func TestNestedElements(t *testing.T) {
	// Test parsing nested elements
	xml := `<outer><inner>Nested content</inner></outer>`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	nodes, ok := doc.DeepFind("inner")
	if !ok {
		t.Fatal("Failed to find inner element")
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 inner node, got %d", len(nodes))
	}

	if nodes[0].GetText() != "Nested content" {
		t.Fatalf("Expected text 'Nested content', got '%s'", nodes[0].GetText())
	}
}

func TestAttributes(t *testing.T) {
	// Test parsing attributes
	xml := `<user name="John" age="30" />`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	nodes, ok := doc.DeepFind("user")
	if !ok {
		t.Fatal("Failed to find user element")
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 user node, got %d", len(nodes))
	}

	nameAttr, ok := nodes[0].GetAttribute("name")
	if !ok {
		t.Fatal("Failed to find name attribute")
	}

	if nameAttr != "John" {
		t.Fatalf("Expected name attribute 'John', got '%s'", nameAttr)
	}

	ageAttr, ok := nodes[0].GetAttribute("age")
	if !ok {
		t.Fatal("Failed to find age attribute")
	}

	if ageAttr != "30" {
		t.Fatalf("Expected age attribute '30', got '%s'", ageAttr)
	}
}

func TestMixedContent(t *testing.T) {
	// Test mixed content (text and elements)
	xml := `Text before <tag>Inside tag</tag> Text after`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(doc.Root.Children) != 3 {
		t.Fatalf("Expected 3 children of root, got %d", len(doc.Root.Children))
	}

	if doc.Root.Children[0].Type != TextNode || doc.Root.Children[0].Value != "Text before " {
		t.Fatalf("Expected first child to be text 'Text before ', got '%s'", doc.Root.Children[0].Value)
	}

	if doc.Root.Children[1].Type != ElementNode || doc.Root.Children[1].Name != "tag" {
		t.Fatal("Expected second child to be tag element")
	}

	if doc.Root.Children[2].Type != TextNode || doc.Root.Children[2].Value != " Text after" {
		t.Fatalf("Expected third child to be text ' Text after', got '%s'", doc.Root.Children[2].Value)
	}
}

func TestComment(t *testing.T) {
	// Test parsing comments
	xml := `<!-- This is a comment --><data>Content</data>`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(doc.Root.Children) != 2 {
		t.Fatalf("Expected 2 children of root, got %d", len(doc.Root.Children))
	}

	if doc.Root.Children[0].Type != CommentNode || doc.Root.Children[0].Value != " This is a comment " {
		t.Fatalf("Expected first child to be comment ' This is a comment ', got '%s'", doc.Root.Children[0].Value)
	}

	nodes, ok := doc.DeepFind("data")
	if !ok {
		t.Fatal("Failed to find data element")
	}

	if nodes[0].GetText() != "Content" {
		t.Fatalf("Expected text 'Content', got '%s'", nodes[0].GetText())
	}
}

func TestProcessingInstruction(t *testing.T) {
	// Test parsing processing instructions
	xml := `<?xml version="1.0" encoding="UTF-8"?><data>Content</data>`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(doc.Root.Children) != 2 {
		t.Fatalf("Expected 2 children of root, got %d", len(doc.Root.Children))
	}

	if doc.Root.Children[0].Type != ProcessingInstructionNode || doc.Root.Children[0].Name != "xml" {
		t.Fatal("Expected first child to be xml processing instruction")
	}

	nodes, ok := doc.DeepFind("data")
	if !ok {
		t.Fatal("Failed to find data element")
	}

	if nodes[0].GetText() != "Content" {
		t.Fatalf("Expected text 'Content', got '%s'", nodes[0].GetText())
	}
}

func TestUnclosedTags(t *testing.T) {
	// Test unclosed tags (where outer closes before inner)
	xml := `<outer><inner>Content</outer>`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	nodes, ok := doc.DeepFind("inner")
	if !ok {
		t.Fatal("Failed to find inner element")
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 inner node, got %d", len(nodes))
	}

	if nodes[0].GetText() != "Content" {
		t.Fatalf("Expected text 'Content', got '%s'", nodes[0].GetText())
	}
}

func TestFindOne(t *testing.T) {
	// Test FindOne method
	xml := `<data><item>First</item><item>Second</item></data>`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	node, ok := doc.FindOne("item")
	if !ok {
		t.Fatal("Failed to find item element")
	}

	if node.GetText() != "First" {
		t.Fatalf("Expected text 'First', got '%s'", node.GetText())
	}
}

func TestPerformance(t *testing.T) {
	// Test performance with a large XML document
	const elementCount = 1000

	// Generate a large XML document
	var xml strings.Builder
	xml.WriteString("<root>")
	for i := 0; i < elementCount; i++ {
		xml.WriteString("<item id=\"")
		xml.WriteByte('A' + byte(i%26))
		xml.WriteString("\">Item ")
		xml.WriteByte('A' + byte(i%26))
		xml.WriteString("</item>")
	}
	xml.WriteString("</root>")

	doc, err := Parse(xml.String())
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	nodes, ok := doc.DeepFind("item")
	if !ok {
		t.Fatal("Failed to find item elements")
	}

	if len(nodes) != elementCount {
		t.Fatalf("Expected %d item nodes, got %d", elementCount, len(nodes))
	}
}

func TestMultipleRootElements(t *testing.T) {
	// Test with multiple root elements (invalid in standard XML)
	xml := `<first>Element 1</first><second>Element 2</second>`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Check both roots are found
	first, ok := doc.DeepFind("first")
	if !ok || len(first) != 1 {
		t.Fatal("Failed to find first element")
	}

	second, ok := doc.DeepFind("second")
	if !ok || len(second) != 1 {
		t.Fatal("Failed to find second element")
	}

	if first[0].GetText() != "Element 1" {
		t.Fatalf("Expected text 'Element 1', got '%s'", first[0].GetText())
	}

	if second[0].GetText() != "Element 2" {
		t.Fatalf("Expected text 'Element 2', got '%s'", second[0].GetText())
	}
}

func TestMalformedAttributes(t *testing.T) {
	// Test with malformed attributes
	xml := `<element attr1 attr2=value attr3="quoted value" attr4='single quotes'>`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	elements, ok := doc.DeepFind("element")
	if !ok || len(elements) != 1 {
		t.Fatal("Failed to find element")
	}

	element := elements[0]

	// Check attribute values
	value, ok := element.GetAttribute("attr1")
	if !ok {
		t.Fatal("Failed to find attr1")
	}
	if value != "" {
		t.Fatalf("Expected empty value for attr1, got '%s'", value)
	}

	value, ok = element.GetAttribute("attr2")
	if !ok {
		t.Fatal("Failed to find attr2")
	}
	if value != "value" {
		t.Fatalf("Expected 'value' for attr2, got '%s'", value)
	}

	value, ok = element.GetAttribute("attr3")
	if !ok {
		t.Fatal("Failed to find attr3")
	}
	if value != "quoted value" {
		t.Fatalf("Expected 'quoted value' for attr3, got '%s'", value)
	}

	value, ok = element.GetAttribute("attr4")
	if !ok {
		t.Fatal("Failed to find attr4")
	}
	if value != "single quotes" {
		t.Fatalf("Expected 'single quotes' for attr4, got '%s'", value)
	}
}

func TestSelfClosingTags(t *testing.T) {
	// Test self-closing tags
	xml := `<parent><selfClosing/><regular>Content</regular></parent>`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	parent, ok := doc.DeepFind("parent")
	if !ok || len(parent) != 1 {
		t.Fatal("Failed to find parent element")
	}

	if len(parent[0].Children) != 2 {
		t.Fatalf("Expected 2 children of parent, got %d", len(parent[0].Children))
	}

	if parent[0].Children[0].Type != ElementNode || parent[0].Children[0].Name != "selfClosing" {
		t.Fatal("Expected first child to be selfClosing element")
	}

	if len(parent[0].Children[0].Children) != 0 {
		t.Fatalf("Expected selfClosing to have no children, got %d", len(parent[0].Children[0].Children))
	}
}

func TestUnexpectedEndOfInput(t *testing.T) {
	// Test with unexpected end of input in various states
	testCases := []string{
		"<",
		"<tag",
		"<tag attribute=",
		"<tag attribute=\"value",
		"<!-- comment",
		"<?xml",
	}

	for _, testCase := range testCases {
		doc, _ := Parse(testCase)

		// Even with errors, we should get a document back
		if doc == nil {
			t.Fatalf("Expected document for input %q, got nil", testCase)
		}
	}
}

func TestDeepNestedElements(t *testing.T) {
	// Test deeply nested elements
	xml := `<level1><level2><level3><level4><level5>Deep</level5></level4></level3></level2></level1>`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	level5, ok := doc.DeepFind("level5")
	if !ok || len(level5) != 1 {
		t.Fatal("Failed to find deeply nested element")
	}

	if level5[0].GetText() != "Deep" {
		t.Fatalf("Expected text 'Deep', got '%s'", level5[0].GetText())
	}
}

func TestLLMUseCase(t *testing.T) {
	// Test the LLM use case
	xml := `<think>
To find the sum of the first 100 positive integers, I can use the formula:
Sum = n(n+1)/2, where n is the number of integers.

In this case, n = 100.

Substituting:
Sum = 100(100+1)/2
Sum = 100(101)/2
Sum = 10100/2
Sum = 5050

Let me double-check this. Another approach would be to pair numbers from opposite ends:
1 + 100 = 101
2 + 99 = 101
3 + 98 = 101
...
50 + 51 = 101

There are 50 such pairs, and each pair sums to 101.
So, total sum = 50 × 101 = 5050

Both methods yield the same result, so I'm confident the answer is 5050.
</think>

<output>The sum of the first 100 positive integers is 5050.

This can be calculated using the formula Sum = n(n+1)/2, where n is the number of integers we're adding.
For n = 100, we get:
Sum = 100(100+1)/2 = 100(101)/2 = 10100/2 = 5050</output>`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	thinking, ok := doc.DeepFind("think")
	if !ok || len(thinking) != 1 {
		t.Fatal("Failed to find thinking element")
	}

	output, ok := doc.DeepFind("output")
	if !ok || len(output) != 1 {
		t.Fatal("Failed to find output element")
	}
}

func TestStringMethod(t *testing.T) {
	// Test the String method for serialization
	xml := `<root><child attribute="value">Content</child></root>`

	doc, err := Parse(xml)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// The string representation should be parseable again
	repr := doc.String()
	doc2, err := Parse(repr)
	if err != nil {
		t.Fatalf("Failed to parse string representation: %v", err)
	}

	child, ok := doc2.DeepFind("child")
	if !ok || len(child) != 1 {
		t.Fatal("Failed to find child element in reparsed document")
	}

	attrValue, ok := child[0].GetAttribute("attribute")
	if !ok || attrValue != "value" {
		t.Fatalf("Expected attribute 'value', got '%s'", attrValue)
	}

	if child[0].GetText() != "Content" {
		t.Fatalf("Expected text 'Content', got '%s'", child[0].GetText())
	}
}
