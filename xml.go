// Package flexml provides a flexible XML parser that can handle partial and invalid XML.
package flexml

import (
	"fmt"
	"strings"
)

// NodeType represents the type of a Node
type NodeType int

const (
	// ElementNode represents an XML element
	ElementNode NodeType = iota
	// TextNode represents text content
	TextNode
	// CommentNode represents an XML comment
	CommentNode
	// ProcessingInstructionNode represents an XML processing instruction
	ProcessingInstructionNode
)

// Node represents an XML node
type Node struct {
	Type     NodeType
	Name     string // Element name or PI target
	Value    string // Text content or PI data
	Children []*Node
	Attrs    map[string]string
	Parent   *Node
}

func (n *Node) FindOne(name string) (*Node, bool) {
	var result []*Node
	deepFindRecursive(n, name, &result)
	if len(result) > 0 {
		return result[0], true
	}
	return nil, false
}

func (n *Node) FindDeep(name string) ([]*Node, bool) {
	var result []*Node
	deepFindRecursive(n, name, &result)
	if len(result) > 0 {
		return result, true
	}
	return nil, false
}

// Document represents an XML document
type Document struct {
	Root *Node
}

// Parse parses an XML string and returns a Document
func Parse(xml string) (*Document, error) {
	parser := &parser{
		input: []byte(xml),
		pos:   0,
		line:  1,
		col:   1,
	}

	doc := &Document{
		Root: &Node{
			Type:     ElementNode,
			Name:     "root", // Special root node to hold everything
			Children: []*Node{},
			Attrs:    map[string]string{},
		},
	}

	if err := parser.parse(doc.Root); err != nil {
		return doc, err // Return partial document with error
	}

	return doc, nil
}

// parser represents the parsing state
type parser struct {
	input    []byte
	pos      int
	line     int
	col      int
	lastChar byte
}

// parse parses XML content and adds nodes to the given parent
func (p *parser) parse(parent *Node) error {
	for p.pos < len(p.input) {
		// Check for tag start
		if p.pos < len(p.input) && p.input[p.pos] == '<' {
			p.advance() // Skip '<'

			// Check what kind of tag we have
			if p.pos < len(p.input) {
				switch p.input[p.pos] {
				case '/': // Closing tag
					p.advance() // Skip '/'
					name, err := p.readName()
					if err != nil {
						return err
					}

					// Skip to end of tag
					for p.pos < len(p.input) && p.input[p.pos] != '>' {
						p.advance()
					}

					if p.pos < len(p.input) {
						p.advance() // Skip '>'
					}

					// Check if this closes our current node
					if parent.Name == name {
						return nil // Successfully closed this node
					}

					// Otherwise, just ignore the closing tag (flexible parsing)

				case '!': // Comment or DOCTYPE
					p.advance() // Skip '!'

					if p.pos+1 < len(p.input) && p.input[p.pos] == '-' && p.input[p.pos+1] == '-' {
						// Comment
						p.advance() // Skip first '-'
						p.advance() // Skip second '-'

						comment, err := p.readUntil("-->")
						if err != nil {
							return err
						}

						commentNode := &Node{
							Type:   CommentNode,
							Value:  comment,
							Parent: parent,
						}

						parent.Children = append(parent.Children, commentNode)
					} else {
						// DOCTYPE or other declaration - treat as text for flexibility
						text := "<!" + p.readUntilChar('>')
						if p.pos < len(p.input) {
							text += string(p.input[p.pos])
							p.advance() // Skip '>'
						}

						textNode := &Node{
							Type:   TextNode,
							Value:  text,
							Parent: parent,
						}

						parent.Children = append(parent.Children, textNode)
					}

				case '?': // Processing instruction
					p.advance() // Skip '?'

					target, err := p.readName()
					if err != nil {
						return err
					}

					// Read PI data
					data, err := p.readUntil("?>")
					if err != nil {
						return err
					}

					piNode := &Node{
						Type:   ProcessingInstructionNode,
						Name:   target,
						Value:  strings.TrimSpace(data),
						Parent: parent,
					}

					parent.Children = append(parent.Children, piNode)

				default: // Opening tag
					name, err := p.readName()
					if err != nil {
						return err
					}

					node := &Node{
						Type:     ElementNode,
						Name:     name,
						Children: []*Node{},
						Attrs:    map[string]string{},
						Parent:   parent,
					}

					// Parse attributes
					for p.pos < len(p.input) && p.input[p.pos] != '>' && p.input[p.pos] != '/' {
						p.skipWhitespace()

						if p.pos < len(p.input) && p.input[p.pos] != '>' && p.input[p.pos] != '/' {
							attrName, attrValue, err := p.readAttribute()
							if err != nil {
								// Treat malformed attribute as end of attributes
								break
							}

							node.Attrs[attrName] = attrValue
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

					// Add node to parent
					parent.Children = append(parent.Children, node)

					// Parse children if not self-closing
					if !selfClosing {
						if err := p.parse(node); err != nil {
							// Just ignore errors when parsing children for flexibility
						}
					}
				}
			} else {
				// End of input after '<', treat as text
				textNode := &Node{
					Type:   TextNode,
					Value:  "<",
					Parent: parent,
				}

				parent.Children = append(parent.Children, textNode)
			}
		} else {
			// Text content
			text := p.readUntilChar('<')

			if text != "" {
				textNode := &Node{
					Type:   TextNode,
					Value:  text,
					Parent: parent,
				}

				parent.Children = append(parent.Children, textNode)
			}
		}
	}

	return nil // Reached end of input
}

// advance moves the parser position forward by one character
func (p *parser) advance() {
	if p.pos < len(p.input) {
		if p.input[p.pos] == '\n' {
			p.line++
			p.col = 1
		} else {
			p.col++
		}

		p.lastChar = p.input[p.pos]
		p.pos++
	}
}

// readName reads an XML name
func (p *parser) readName() (string, error) {
	p.skipWhitespace()

	nameStart := p.pos

	// First character must be a letter, underscore or colon
	if p.pos < len(p.input) {
		ch := p.input[p.pos]
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == ':') {
			return "", fmt.Errorf("invalid name start character at line %d, column %d", p.line, p.col)
		}

		p.advance()
	} else {
		return "", fmt.Errorf("unexpected end of input when reading name at line %d, column %d", p.line, p.col)
	}

	// Subsequent characters can include digits, hyphens, periods
	for p.pos < len(p.input) {
		ch := p.input[p.pos]

		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') ||
			ch == '_' || ch == ':' || ch == '-' || ch == '.' {
			p.advance()
		} else {
			break
		}
	}

	if p.pos > nameStart {
		return string(p.input[nameStart:p.pos]), nil
	}

	return "", fmt.Errorf("empty name at line %d, column %d", p.line, p.col)
}

// readAttribute reads an attribute name and value
func (p *parser) readAttribute() (string, string, error) {
	name, err := p.readName()
	if err != nil {
		return "", "", err
	}

	p.skipWhitespace()

	// Check for equals sign
	if p.pos >= len(p.input) || p.input[p.pos] != '=' {
		// For flexibility, allow attributes without values
		return name, "", nil
	}

	p.advance() // Skip '='
	p.skipWhitespace()

	// Read value
	if p.pos >= len(p.input) {
		return name, "", nil // Empty value for flexibility
	}

	if p.input[p.pos] == '"' || p.input[p.pos] == '\'' {
		quote := p.input[p.pos]
		p.advance() // Skip quote

		valueStart := p.pos

		for p.pos < len(p.input) && p.input[p.pos] != quote {
			p.advance()
		}

		value := string(p.input[valueStart:p.pos])

		if p.pos < len(p.input) {
			p.advance() // Skip closing quote
		}

		return name, value, nil
	} else {
		// Unquoted value (non-standard but flexible)
		valueStart := p.pos

		for p.pos < len(p.input) && !isWhitespace(p.input[p.pos]) && p.input[p.pos] != '>' && p.input[p.pos] != '/' {
			p.advance()
		}

		value := string(p.input[valueStart:p.pos])
		return name, value, nil
	}
}

// readUntil reads until the given delimiter is found
func (p *parser) readUntil(delimiter string) (string, error) {
	start := p.pos

	for p.pos <= len(p.input)-len(delimiter) {
		if string(p.input[p.pos:p.pos+len(delimiter)]) == delimiter {
			result := string(p.input[start:p.pos])

			// Advance past delimiter
			for i := 0; i < len(delimiter); i++ {
				p.advance()
			}

			return result, nil
		}

		p.advance()
	}

	// Reached end of input without finding delimiter
	result := string(p.input[start:p.pos])
	return result, fmt.Errorf("unexpected end of input while looking for %q", delimiter)
}

// readUntilChar reads until the given character is found
func (p *parser) readUntilChar(ch byte) string {
	start := p.pos

	for p.pos < len(p.input) && p.input[p.pos] != ch {
		p.advance()
	}

	return string(p.input[start:p.pos])
}

// skipWhitespace skips whitespace characters
func (p *parser) skipWhitespace() {
	for p.pos < len(p.input) && isWhitespace(p.input[p.pos]) {
		p.advance()
	}
}

// isWhitespace checks if a byte is a whitespace character
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// DeepFind searches for nodes with the given name, recursively
func (d *Document) DeepFind(name string) ([]*Node, bool) {
	var result []*Node
	deepFindRecursive(d.Root, name, &result)
	return result, len(result) > 0
}

// Helper function for recursive search
func deepFindRecursive(node *Node, name string, result *[]*Node) {
	if node.Type == ElementNode && node.Name == name {
		*result = append(*result, node)
	}

	for _, child := range node.Children {
		deepFindRecursive(child, name, result)
	}
}

// FindOne finds the first node with the given name
func (d *Document) FindOne(name string) (*Node, bool) {
	nodes, found := d.DeepFind(name)
	if found && len(nodes) > 0 {
		return nodes[0], true
	}
	return nil, false
}

// GetAttribute returns the value of an attribute
func (n *Node) GetAttribute(name string) (string, bool) {
	val, ok := n.Attrs[name]
	return val, ok
}

// GetText returns the text content of a node (concatenating all text child nodes)
func (n *Node) GetText() string {
	var sb strings.Builder

	for _, child := range n.Children {
		if child.Type == TextNode {
			sb.WriteString(child.Value)
		} else if child.Type == ElementNode {
			// Recursively get text from child elements
			sb.WriteString(child.GetText())
		}
	}

	return sb.String()
}

// String returns a string representation of the document
func (d *Document) String() string {
	var sb strings.Builder
	printNode(&sb, d.Root, 0)
	return sb.String()
}

// Helper function to print a node recursively
func printNode(sb *strings.Builder, node *Node, indent int) {
	indentStr := strings.Repeat("  ", indent)

	switch node.Type {
	case ElementNode:
		sb.WriteString(indentStr)
		sb.WriteString("<")
		sb.WriteString(node.Name)

		for k, v := range node.Attrs {
			sb.WriteString(" ")
			sb.WriteString(k)
			sb.WriteString("=\"")
			sb.WriteString(v)
			sb.WriteString("\"")
		}

		if len(node.Children) == 0 {
			sb.WriteString("/>")
		} else {
			sb.WriteString(">")

			hasChildElements := false
			for _, child := range node.Children {
				if child.Type == ElementNode {
					hasChildElements = true
					sb.WriteString("\n")
					printNode(sb, child, indent+1)
				} else {
					printNode(sb, child, 0)
				}
			}

			if hasChildElements {
				sb.WriteString("\n")
				sb.WriteString(indentStr)
			}

			sb.WriteString("</")
			sb.WriteString(node.Name)
			sb.WriteString(">")
		}

	case TextNode:
		sb.WriteString(node.Value)

	case CommentNode:
		sb.WriteString(indentStr)
		sb.WriteString("<!--")
		sb.WriteString(node.Value)
		sb.WriteString("-->")

	case ProcessingInstructionNode:
		sb.WriteString(indentStr)
		sb.WriteString("<?")
		sb.WriteString(node.Name)
		if node.Value != "" {
			sb.WriteString(" ")
			sb.WriteString(node.Value)
		}
		sb.WriteString("?>")
	}
}
