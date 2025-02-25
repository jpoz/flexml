// Package flexml provides a flexible XML parser that can handle partial and invalid XML.
package flexml

import (
	"io"
	"strings"
)

// EventType represents the type of XML event
type EventType int

const (
	// StartElement represents the start of an XML element
	StartElement EventType = iota
	// EndElement represents the end of an XML element
	EndElement
	// Text represents text content
	Text
	// Comment represents an XML comment
	Comment
	// ProcessingInstruction represents an XML processing instruction
	ProcessingInstruction
)

// Event represents an XML parsing event
type Event struct {
	Type        EventType
	Name        string            // Element name or PI target
	Text        string            // Text content, comment, or PI data
	Attributes  map[string]string // Element attributes
	SelfClosing bool              // Whether the element is self-closing
}

// Stream represents an XML parser that processes input in a streaming fashion
type Stream struct {
	parser       *parser
	buffer       []byte
	position     int
	currentEvent *Event
	err          error
}

// NewStream creates a new XML stream parser
func NewStream() *Stream {
	return &Stream{
		parser: &parser{
			pos:  0,
			line: 1,
			col:  1,
		},
		buffer:   make([]byte, 0),
		position: 0,
	}
}

// AddData adds more data to the stream parser
func (s *Stream) AddData(data []byte) {
	if len(s.buffer) == 0 {
		s.buffer = data
		s.parser.input = s.buffer
	} else {
		// Append new data
		s.buffer = append(s.buffer, data...)
		// Update parser input
		s.parser.input = s.buffer
	}
}

// Next advances to the next event
func (s *Stream) Next() bool {
	if s.err != nil || s.position >= len(s.buffer) {
		return false
	}

	s.parser.pos = s.position
	event, newPos, err := s.parser.nextEvent()
	s.currentEvent = event
	s.position = newPos
	s.err = err

	return event != nil
}

// Event returns the current event
func (s *Stream) Event() *Event {
	return s.currentEvent
}

// Err returns any error that occurred during parsing
func (s *Stream) Err() error {
	return s.err
}

// Parse parses an XML stream from an io.Reader
func ParseStream(r io.Reader) (*Stream, error) {
	stream := NewStream()

	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			stream.AddData(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return stream, err
		}
	}

	return stream, nil
}

// nextEvent parses the next XML event
func (p *parser) nextEvent() (*Event, int, error) {
	// Skip any whitespace
	p.skipWhitespace()

	// Check if we've reached the end of input
	if p.pos >= len(p.input) {
		return nil, p.pos, nil
	}

	// Check for tag start
	if p.input[p.pos] == '<' {
		p.advance() // Skip '<'

		// Check what kind of tag we have
		if p.pos < len(p.input) {
			switch p.input[p.pos] {
			case '/': // Closing tag
				p.advance() // Skip '/'
				name, err := p.readName()
				if err != nil {
					return nil, p.pos, err
				}

				// Skip to end of tag
				for p.pos < len(p.input) && p.input[p.pos] != '>' {
					p.advance()
				}

				if p.pos < len(p.input) {
					p.advance() // Skip '>'
				}

				return &Event{
					Type: EndElement,
					Name: name,
				}, p.pos, nil

			case '!': // Comment or DOCTYPE
				p.advance() // Skip '!'

				if p.pos+1 < len(p.input) && p.input[p.pos] == '-' && p.input[p.pos+1] == '-' {
					// Comment
					p.advance() // Skip first '-'
					p.advance() // Skip second '-'

					comment, err := p.readUntil("-->")
					if err != nil {
						return nil, p.pos, err
					}

					return &Event{
						Type: Comment,
						Text: comment,
					}, p.pos, nil
				} else {
					// DOCTYPE or other declaration - treat as text for flexibility
					text := "<!" + p.readUntilChar('>')
					if p.pos < len(p.input) {
						text += string(p.input[p.pos])
						p.advance() // Skip '>'
					}

					return &Event{
						Type: Text,
						Text: text,
					}, p.pos, nil
				}

			case '?': // Processing instruction
				p.advance() // Skip '?'

				target, err := p.readName()
				if err != nil {
					return nil, p.pos, err
				}

				// Read PI data
				data, err := p.readUntil("?>")
				if err != nil {
					return nil, p.pos, err
				}

				return &Event{
					Type: ProcessingInstruction,
					Name: target,
					Text: data,
				}, p.pos, nil

			default: // Opening tag
				name, err := p.readName()
				if err != nil {
					return nil, p.pos, err
				}

				attrs := make(map[string]string)

				// Parse attributes
				for p.pos < len(p.input) && p.input[p.pos] != '>' && p.input[p.pos] != '/' {
					p.skipWhitespace()

					if p.pos < len(p.input) && p.input[p.pos] != '>' && p.input[p.pos] != '/' {
						attrName, attrValue, err := p.readAttribute()
						if err != nil {
							// Treat malformed attribute as end of attributes
							break
						}

						attrs[attrName] = attrValue
					}
				}

				// Check for self-closing tag
				selfClosing := false
				if p.pos < len(p.input) && p.input[p.pos] == '/' {
					selfClosing = true
					p.advance() // Skip '/'
				}

				// Skip to end of tag
				if p.pos < len(p.input) && p.input[p.pos] == '>' {
					p.advance() // Skip '>'
				}

				return &Event{
					Type:        StartElement,
					Name:        name,
					Attributes:  attrs,
					SelfClosing: selfClosing,
				}, p.pos, nil
			}
		} else {
			// End of input after '<', treat as text
			return &Event{
				Type: Text,
				Text: "<",
			}, p.pos, nil
		}
	} else {
		// Text content
		text := p.readUntilChar('<')

		if text != "" {
			return &Event{
				Type: Text,
				Text: text,
			}, p.pos, nil
		}
	}

	return nil, p.pos, nil
}

// NewElementStreamReader creates a new reader for XML stream events
func NewElementStreamReader(r io.Reader) *ElementStreamReader {
	stream := NewStream()

	return &ElementStreamReader{
		reader: r,
		stream: stream,
		buffer: make([]byte, 4096),
	}
}

// ElementStreamReader reads XML and produces events
type ElementStreamReader struct {
	reader      io.Reader
	stream      *Stream
	buffer      []byte
	currentNode *Node
	stack       []*Node
	eof         bool
}

// ReadNode reads the next complete XML node
func (e *ElementStreamReader) ReadNode() (*Node, error) {
	// If we haven't loaded any data yet, read the first chunk
	if len(e.stream.buffer) == 0 && !e.eof {
		if err := e.readMoreData(); err != nil && err != io.EOF {
			return nil, err
		}
	}

	// Loop through events to build a complete node
	for e.stream.Next() {
		event := e.stream.Event()

		switch event.Type {
		case StartElement:
			node := &Node{
				Type:     ElementNode,
				Name:     event.Name,
				Children: []*Node{},
				Attrs:    event.Attributes,
			}

			if len(e.stack) == 0 {
				// This is a root node
				if event.SelfClosing {
					return node, nil
				}

				e.currentNode = node
				e.stack = append(e.stack, node)
			} else {
				// Add as child to current node
				parent := e.stack[len(e.stack)-1]
				node.Parent = parent
				parent.Children = append(parent.Children, node)

				if !event.SelfClosing {
					e.stack = append(e.stack, node)
				}
			}

		case EndElement:
			if len(e.stack) > 0 {
				// Pop the stack
				e.stack = e.stack[:len(e.stack)-1]

				// If this completes a root node, return it
				if len(e.stack) == 0 {
					result := e.currentNode
					e.currentNode = nil
					return result, nil
				}
			}

		case Text:
			if len(e.stack) > 0 {
				parent := e.stack[len(e.stack)-1]
				textNode := &Node{
					Type:   TextNode,
					Value:  event.Text,
					Parent: parent,
				}
				parent.Children = append(parent.Children, textNode)
			}

		case Comment:
			if len(e.stack) > 0 {
				parent := e.stack[len(e.stack)-1]
				commentNode := &Node{
					Type:   CommentNode,
					Value:  event.Text,
					Parent: parent,
				}
				parent.Children = append(parent.Children, commentNode)
			}

		case ProcessingInstruction:
			if len(e.stack) > 0 {
				parent := e.stack[len(e.stack)-1]
				piNode := &Node{
					Type:   ProcessingInstructionNode,
					Name:   event.Name,
					Value:  event.Text,
					Parent: parent,
				}
				parent.Children = append(parent.Children, piNode)
			}
		}

		// Check if we need more data
		if !e.eof && e.stream.position >= len(e.stream.buffer)-1024 {
			if err := e.readMoreData(); err != nil && err != io.EOF {
				return nil, err
			}
		}
	}

	// If we have a current node but hit the end of input, return it anyway
	if e.currentNode != nil {
		result := e.currentNode
		e.currentNode = nil
		e.stack = nil
		return result, nil
	}

	// Check if we need more data and haven't reached EOF
	if !e.eof {
		if err := e.readMoreData(); err != nil {
			if err == io.EOF {
				e.eof = true
				return nil, io.EOF
			}
			return nil, err
		}
		return e.ReadNode()
	}

	return nil, io.EOF
}

// ReadMoreData reads more data from the underlying reader
func (e *ElementStreamReader) readMoreData() error {
	n, err := e.reader.Read(e.buffer)
	if n > 0 {
		e.stream.AddData(e.buffer[:n])
	}
	if err == io.EOF {
		e.eof = true
	}
	return err
}

// NewStreamDocument creates a new StreamDocument to collect XML nodes
func NewStreamDocument() *StreamDocument {
	return &StreamDocument{
		Nodes: []*Node{},
	}
}

// StreamDocument represents a collection of streamed XML nodes
type StreamDocument struct {
	Nodes []*Node
}

// AddNode adds a node to the document
func (d *StreamDocument) AddNode(node *Node) {
	d.Nodes = append(d.Nodes, node)
}

// DeepFind searches for nodes with the given name, recursively
func (d *StreamDocument) DeepFind(name string) ([]*Node, bool) {
	var result []*Node

	for _, node := range d.Nodes {
		deepFindRecursive(node, name, &result)
	}

	return result, len(result) > 0
}

// FindOne finds the first node with the given name
func (d *StreamDocument) FindOne(name string) (*Node, bool) {
	nodes, found := d.DeepFind(name)
	if found && len(nodes) > 0 {
		return nodes[0], true
	}
	return nil, false
}

// String returns a string representation of the document
func (d *StreamDocument) String() string {
	var sb strings.Builder

	for _, node := range d.Nodes {
		printNode(&sb, node, 0)
		sb.WriteString("\n")
	}

	return sb.String()
}

// ParseReader parses XML from an io.Reader and returns a StreamDocument
func ParseReader(r io.Reader) (*StreamDocument, error) {
	// Create a raw stream for more direct processing of events
	stream, err := ParseStream(r)
	if err != nil {
		return nil, err
	}

	doc := NewStreamDocument()

	// For root-level nodes we'll collect text and elements directly
	var currentNode *Node = nil

	for stream.Next() {
		event := stream.Event()

		switch event.Type {
		case StartElement:
			// Create a new node
			node := &Node{
				Type:     ElementNode,
				Name:     event.Name,
				Children: []*Node{},
				Attrs:    event.Attributes,
			}

			if currentNode == nil {
				// This is a root node
				if !event.SelfClosing {
					// Process this node and its children fully
					subReader := strings.NewReader("") // Empty reader
					subStream := NewElementStreamReader(subReader)
					subStream.currentNode = node    // Set this as the current node
					subStream.stack = []*Node{node} // Start with this node on the stack

					// Manually process remaining events until this element closes
					depth := 1
					for stream.Next() && depth > 0 {
						subEvent := stream.Event()

						switch subEvent.Type {
						case StartElement:
							subNode := &Node{
								Type:     ElementNode,
								Name:     subEvent.Name,
								Children: []*Node{},
								Attrs:    subEvent.Attributes,
								Parent:   node,
							}

							node.Children = append(node.Children, subNode)
							if !subEvent.SelfClosing {
								node = subNode // Navigate down
								depth++
							}

						case EndElement:
							if depth > 1 {
								// Find parent
								parent := node.Parent
								node = parent // Navigate up
							}
							depth--

						case Text:
							textNode := &Node{
								Type:   TextNode,
								Value:  subEvent.Text,
								Parent: node,
							}
							node.Children = append(node.Children, textNode)

						case Comment:
							commentNode := &Node{
								Type:   CommentNode,
								Value:  subEvent.Text,
								Parent: node,
							}
							node.Children = append(node.Children, commentNode)

						case ProcessingInstruction:
							piNode := &Node{
								Type:   ProcessingInstructionNode,
								Name:   subEvent.Name,
								Value:  subEvent.Text,
								Parent: node,
							}
							node.Children = append(node.Children, piNode)
						}
					}

					// Add the completed node
					doc.AddNode(subStream.currentNode)
				} else {
					// Self-closing root node
					doc.AddNode(node)
				}
			}

		case Text:
			// Add text as a root node
			textNode := &Node{
				Type:  TextNode,
				Value: event.Text,
			}
			doc.AddNode(textNode)

		case Comment:
			// Add comment as a root node
			commentNode := &Node{
				Type:  CommentNode,
				Value: event.Text,
			}
			doc.AddNode(commentNode)

		case ProcessingInstruction:
			// Add PI as a root node
			piNode := &Node{
				Type:  ProcessingInstructionNode,
				Name:  event.Name,
				Value: event.Text,
			}
			doc.AddNode(piNode)
		}
	}

	return doc, nil
}

