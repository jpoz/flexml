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

### Processing LLM Reasoning with `<think>` Tags

LLMs (Large Language Models) often use XML-like tags to structure their reasoning process. FleXML is perfect for parsing this output, extracting both the reasoning steps and the final answer:

```go
package main

import (
    "fmt"
    "strings"
    "github.com/jpoz/flexml"
)

func main() {
    // Example LLM output with reasoning in <think> tags
    llmOutput := `<think>
To solve this problem, I'll need to find the sum of the first 100 positive integers.

I can use the formula: sum = n(n+1)/2
Where n is the number of integers we're adding.

For n = 100:
sum = 100(100+1)/2
sum = 100(101)/2
sum = 10100/2
sum = 5050
</think>

<answer>The sum of the first 100 positive integers is 5050.

I can find this using the formula sum = n(n+1)/2, where n is the number of integers.
For n = 100: sum = 100(100+1)/2 = 100(101)/2 = 10100/2 = 5050</answer>`

    // Parse the LLM output
    doc, err := flexml.Parse(llmOutput)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    // Extract the thinking process
    thinkNode, found := doc.FindOne("think")
    if found {
        thinking := strings.TrimSpace(thinkNode.GetText())
        fmt.Println("=== LLM Reasoning Process ===")
        fmt.Println(thinking)
        fmt.Println("============================")
        fmt.Println()
    }
    
    // Extract the answer
    answerNode, found := doc.FindOne("answer")
    if found {
        answer := strings.TrimSpace(answerNode.GetText())
        fmt.Println("=== Final Answer ===")
        fmt.Println(answer)
        fmt.Println("===================")
    }
    
    // You can also process the reasoning steps for additional analysis
    // For example, to extract mathematical calculations, specific reasoning steps, etc.
}
```

This example shows how you can use FleXML to:
1. Parse LLM outputs that use XML-like tags to structure their thinking
2. Extract and separate the reasoning process from the final answer
3. Process incomplete or malformed XML that might be generated by LLMs

This is useful for applications that want to expose the LLM's reasoning process for review, educational purposes, or to provide transparency in AI decision-making.

You can also use the streaming API for processing LLM outputs as they arrive:

```go
package main

import (
    "fmt"
    "io"
    "net/http"
    "strings"
    "github.com/jpoz/flexml"
)

func main() {
    // Simulate an API request receiving LLM output in chunks
    // In a real application, this would be replaced with an actual HTTP request
    // to an LLM API endpoint with streaming enabled
    llmOutputChunks := []string{
        "<th",
        "ink>\nLet me reason through this step by step.\n\nTo find",
        " the derivative of f(x) = x^2 * sin(x), I'll use the product rule:\n",
        "(uv)' = u'v + uv'\n\nLet u = x^2 and v = sin(x)\n",
        "Then u' = 2x and v' = cos(x)\n\n",
        "Applying the product rule:\nf'(x) = (2x)(sin(x)) + (x^2)(cos(x))",
        "\n= 2x*sin(x) + x^2*cos(x)",
        "</think>\n\n<answer>",
        "The derivative of f(x) = x^2 * sin(x) is:\n\nf'(x) = 2x*sin(x) + x^2*cos(x)",
        "</answer>"
    }
    
    // Create a stream
    stream := flexml.NewStream()
    
    // Variables to track state
    inThinking := false
    inAnswer := false
    thinking := ""
    answer := ""
    
    // Process the chunks as they arrive
    for _, chunk := range llmOutputChunks {
        // Add the chunk to the stream
        stream.AddData([]byte(chunk))
        
        // Process events from the stream
        for stream.Next() {
            event := stream.Event()
            
            switch event.Type {
            case flexml.StartElement:
                if event.Name == "think" {
                    inThinking = true
                    inAnswer = false
                    fmt.Println("Started receiving thinking process...")
                } else if event.Name == "answer" {
                    inThinking = false
                    inAnswer = true
                    fmt.Println("Started receiving answer...")
                }
                
            case flexml.EndElement:
                if event.Name == "think" {
                    inThinking = false
                    fmt.Println("Completed thinking section.")
                } else if event.Name == "answer" {
                    inAnswer = false
                    fmt.Println("Completed answer section.")
                }
                
            case flexml.Text:
                if inThinking {
                    thinking += event.Text
                    fmt.Printf("Thinking (partial): %s\n", strings.TrimSpace(event.Text))
                } else if inAnswer {
                    answer += event.Text
                    fmt.Printf("Answer (partial): %s\n", strings.TrimSpace(event.Text))
                }
            }
        }
    }
    
    // Signal that we're done adding data
    stream.EOF()
    
    // Process any remaining events
    for stream.Next() {
        // Similar event processing as above
    }
    
    // Final output
    fmt.Println("\n=== Final Thinking Process ===")
    fmt.Println(strings.TrimSpace(thinking))
    fmt.Println("\n=== Final Answer ===")
    fmt.Println(strings.TrimSpace(answer))
}
```

This streaming approach is particularly useful for:
- Processing real-time LLM outputs as they're generated
- Providing immediate feedback to users while the LLM is still thinking
- Working with very large outputs without waiting for the complete response

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
