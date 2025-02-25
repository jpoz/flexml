# FleXML: A Flexible XML Parser for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/jpoz/flexml.svg)](https://pkg.go.dev/github.com/jpoz/flexml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jpoz/flexml)](https://goreportcard.com/report/github.com/jpoz/flexml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

FleXML parses incomplete or invalid XML data. Unlike standard XML parsers that require valid, complete XML input, FleXML gracefully handles partial XML fragments and malformed documents, extracting as much structured data as possible.

## 🌟 Features

- **Partial XML Parsing**: Extract data from incomplete XML fragments
  - `<key>Hello` → Element with text "Hello"
  - `<response><message>` → Nested element structure

- **Invalid XML Handling**: Process XML with mixed content and other issues
  - Handles text mixed with elements
  - Supports malformed or unquoted attributes

- **Flexible Document Querying**: Navigate and extract data easily
  - Find elements by name anywhere in the document
  - Extract text content and attribute values

- **Resilient Parsing**: Recovers gracefully from unexpected input
  - No panic on malformed input
  - Extracts maximum valid data even from corrupted XML

- **Zero Dependencies**: Pure Go implementation with no external dependencies

## 📦 Installation

```bash
go get github.com/jpoz/flexml
```

## 🚀 Quick Start

### Parsing Valid XML

```go
package main

import (
    "fmt"
    "github.com/jpoz/flexml"
)

func main() {
    // Parse complete XML
    xmlStr := `<response><message>Greetings</message></response>`
    
    doc, err := flexml.Parse(xmlStr)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    nodes, ok := doc.DeepFind("message")
    if ok {
        fmt.Printf("Message content: %s\n", nodes[0].GetText())
        // Output: Message content: Greetings
    }
}
```

### Handling Partial XML

```go
package main

import (
    "fmt"
    "github.com/jpoz/flexml"
)

func main() {
    // Example partial XML fragments
    partialXMLs := []string{
        `<key>Hello`,
        `<response><user id="123"`,
        `<data>Value</data></response>`,
    }

    // Process each fragment
    for _, xml := range partialXMLs {
        fmt.Printf("Processing: %s\n", xml)
        
        doc, err := flexml.Parse(xml)
        if err != nil {
            fmt.Printf("Notice (parsing continues): %v\n", err)
        }

        // Even with errors, we can still extract data
        fmt.Printf("Document structure: %s\n", doc.String())
    }
}
```

## ⚙️ How It Works

FleXML uses a custom parsing algorithm designed to be forgiving while still providing a useful document structure:

1. **Single-Pass Parser**: Processes the input in a single pass for efficiency
2. **Node Hierarchy**: Builds a tree of nodes (elements, text, comments)
3. **Automatic Recovery**: Detects and handles common XML errors

The library intelligently handles problematic input by:
- Treating unclosed tags as valid elements
- Supporting text outside of elements
- Processing malformed attributes

## 🔍 API Reference

### Document

- `Parse(xml string) (*Document, error)` - Parses an XML string into a Document
- `DeepFind(name string) ([]*Node, bool)` - Searches for nodes with the given name recursively
- `FindOne(name string) (*Node, bool)` - Finds the first node with the given name
- `String() string` - Returns a string representation of the document

### Node

- `GetAttribute(name string) (string, bool)` - Returns the value of an attribute
- `GetText() string` - Returns the text content of a node
- `Type` - The type of node (ElementNode, TextNode, CommentNode, ProcessingInstructionNode)

## 🧪 Testing

The library includes comprehensive test coverage for both valid and invalid XML parsing:

```bash
go test -v github.com/jpoz/flexml
```
