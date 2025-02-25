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

- **Streaming XML Processing**: Process XML data in chunks or as a stream
  - Parse large XML files without loading them entirely into memory
  - Process XML as it arrives, ideal for network streams
  - Event-based API for efficient processing

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

### Streaming XML Processing

```go
package main

import (
    "fmt"
    "strings"
    "github.com/jpoz/flexml"
)

func main() {
    // Create a stream from a reader
    xmlData := `<users>
        <user id="1">
            <name>Alice</name>
            <email>alice@example.com</email>
        </user>
        <user id="2">
            <name>Bob</name>
            <email>bob@example.com</email>
        </user>
    </users>`
    
    reader := strings.NewReader(xmlData)
    stream, err := flexml.ParseStream(reader)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    // Process events as they are generated
    for stream.Next() {
        event := stream.Event()
        
        switch event.Type {
        case flexml.StartElement:
            fmt.Printf("Element start: %s\n", event.Name)
            if event.Name == "user" {
                if id, ok := event.Attributes["id"]; ok {
                    fmt.Printf("  User ID: %s\n", id)
                }
            }
        case flexml.EndElement:
            fmt.Printf("Element end: %s\n", event.Name)
        case flexml.Text:
            if event.Text != "" {
                fmt.Printf("Text: %s\n", event.Text)
            }
        }
    }
    
    if stream.Err() != nil {
        fmt.Printf("Stream error: %v\n", stream.Err())
    }
}
```

### Reading Complete XML Nodes from Stream

```go
package main

import (
    "fmt"
    "io"
    "os"
    "github.com/jpoz/flexml"
)

func main() {
    // Open a large XML file
    file, err := os.Open("large.xml")
    if err != nil {
        fmt.Printf("Error opening file: %v\n", err)
        return
    }
    defer file.Close()
    
    // Create a node reader
    reader := flexml.NewElementStreamReader(file)
    
    // Read and process complete nodes one at a time
    for {
        node, err := reader.ReadNode()
        if err == io.EOF {
            break
        }
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            continue
        }
        
        // Process the node
        if node.Type == flexml.ElementNode && node.Name == "item" {
            fmt.Printf("Found item: %s\n", node.GetText())
            
            // Check attributes
            if id, ok := node.GetAttribute("id"); ok {
                fmt.Printf("  ID: %s\n", id)
            }
        }
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

### Streaming

- `ParseStream(r io.Reader) (*Stream, error)` - Creates a stream parser from an io.Reader
- `NewStream()` - Creates a new XML stream parser
- `Stream.AddData(data []byte)` - Adds more data to the stream parser
- `Stream.Next() bool` - Advances to the next event, returns false when done
- `Stream.Event() *Event` - Returns the current event
- `Stream.Err() error` - Returns any error that occurred during parsing

### Node Streaming

- `NewElementStreamReader(r io.Reader) *ElementStreamReader` - Creates a reader for XML stream events
- `ElementStreamReader.ReadNode() (*Node, error)` - Reads the next complete XML node
- `ParseReader(r io.Reader) (*StreamDocument, error)` - Parses XML from an io.Reader into a StreamDocument
- `StreamDocument.DeepFind(name string) ([]*Node, bool)` - Searches for nodes in the streamed document
- `StreamDocument.FindOne(name string) (*Node, bool)` - Finds the first matching node in the streamed document

## 🧪 Testing

The library includes comprehensive test coverage for both valid and invalid XML parsing:

```bash
go test -v github.com/jpoz/flexml
```
