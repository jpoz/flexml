# FlexML

[![Go Reference](https://pkg.go.dev/badge/github.com/jpoz/flexml.svg)](https://pkg.go.dev/github.com/jpoz/flexml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jpoz/flexml)](https://goreportcard.com/report/github.com/jpoz/flexml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

FleXML is a flexible XML parser for Go that can handle partial and invalid XML. It's designed to be more forgiving than traditional XML parsers while still providing a useful document structure for querying.

## Features

- Parse partial XML (unclosed tags)
- Parse invalid XML (text mixed with elements)
- Handle malformed or unquoted attributes
- Support for XML comments and processing instructions
- Efficient single-pass parsing algorithm
- Simple API for querying the document

## Installation

```bash
go get github.com/jpoz/flexml
```

## Usage

```go
import "github.com/jpoz/flexml"

// Parse an XML string
doc, err := flexml.Parse("<response><message>Greetings</message></response>")
if err != nil {
    // Handle error (note: even with errors, a partial document may be returned)
}

// Find elements by name using DeepFind
nodes, ok := doc.DeepFind("message")
if ok {
    fmt.Println("Found", len(nodes), "message elements")
    fmt.Println("Text content:", nodes[0].GetText())
}

// Find a single element
node, ok := doc.FindOne("message")
if ok {
    fmt.Println("First message:", node.GetText())
}

// Access attributes
attrValue, ok := node.GetAttribute("id")
if ok {
    fmt.Println("Attribute value:", attrValue)
}

// Get string representation of the document
fmt.Println(doc.String())
```

## Examples

### Parsing valid XML

```go
docStr := `<response><message>Greetings</message></response>`
doc, err := flexml.Parse(docStr)
if err != nil {
    log.Fatalf("Error parsing XML: %v", err)
}

nodes, ok := doc.DeepFind("message")
if ok {
    fmt.Println("Found message:", nodes[0].GetText())
}
```

### Parsing partial XML (unclosed tag)

```go
partialXML := `<key>Hello`
doc, err := flexml.Parse(partialXML)
if err != nil {
    log.Fatalf("Error parsing partial XML: %v", err)
}

nodes, ok := doc.DeepFind("key")
if ok {
    fmt.Println("Found key:", nodes[0].GetText())
}
```

### Parsing invalid XML (text mixed with elements)

```go
invalidXML := `Hello how are you
<name>James</name>`
doc, err := flexml.Parse(invalidXML)
if err != nil {
    log.Fatalf("Error parsing invalid XML: %v", err)
}

// Access text content
if len(doc.Root.Children) > 0 && doc.Root.Children[0].Type == flexml.TextNode {
    fmt.Println("Text content:", doc.Root.Children[0].Value)
}

// Find the name element
nodes, ok := doc.DeepFind("name")
if ok {
    fmt.Println("Found name:", nodes[0].GetText())
}
```

## API Reference

### Document

- `Parse(xml string) (*Document, error)` - Parses an XML string and returns a Document
- `DeepFind(name string) ([]*Node, bool)` - Searches for nodes with the given name recursively
- `FindOne(name string) (*Node, bool)` - Finds the first node with the given name
- `String() string` - Returns a string representation of the document

### Node

- `GetAttribute(name string) (string, bool)` - Returns the value of an attribute
- `GetText() string` - Returns the text content of a node (concatenating all text child nodes)

### Node Types

- `ElementNode` - An XML element node
- `TextNode` - A text node
- `CommentNode` - An XML comment node
- `ProcessingInstructionNode` - An XML processing instruction node

## Performance Considerations

FlexML uses a single-pass parsing algorithm that is generally efficient for most XML documents. However, for extremely large documents, you may want to consider streaming parsers. FlexML's flexibility comes with a slight performance cost compared to strict parsers, but the difference is negligible for most use cases.

## License

MIT License
