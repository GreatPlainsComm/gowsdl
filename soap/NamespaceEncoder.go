package soap

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

type NamespaceEncoder struct {
	encoder    *xml.Encoder
	buffer     *bytes.Buffer
	writer     io.Writer
	namespaces map[string]string
	defaultNS  string
}

func NewNamespaceEncoder(w io.Writer, namespaces map[string]string, defaultNS string) *NamespaceEncoder {
	return newNamespaceEncoder(w, namespaces, defaultNS)
}

func newNamespaceEncoder(w io.Writer, namespaces map[string]string, defaultNS string) *NamespaceEncoder {
	buffer := &bytes.Buffer{}
	return &NamespaceEncoder{
		encoder:    xml.NewEncoder(buffer),
		buffer:     buffer,
		writer:     w,
		namespaces: namespaces,
		defaultNS:  defaultNS,
	}
}

func (ne *NamespaceEncoder) Encode(v interface{}) error {
	// First encode normally
	if err := ne.encoder.Encode(v); err != nil {
		return err
	}
	return nil
}

func (ne *NamespaceEncoder) Flush() error {
	if err := ne.encoder.Flush(); err != nil {
		return err
	}

	// Get the XML string and add namespace prefixes
	xmlStr := ne.buffer.String()

	// Apply namespace prefixes to XML
	xmlStr = ne.applyNamespaces(xmlStr)

	// Write the modified XML to the original writer
	_, err := ne.writer.Write([]byte(xmlStr))
	return err
}

func (ne *NamespaceEncoder) applyNamespaces(xmlStr string) string {
	// Add namespace declarations to root element
	xmlStr = ne.addNamespaceDeclarations(xmlStr)

	// Apply default namespace prefix to non-standard elements
	if ne.defaultNS != "" {
		xmlStr = ne.addDefaultNamespacePrefix(xmlStr)
	} else {
		// Even without default namespace, check for xsi: attributes
		xmlStr = ne.addXsiNamespaceDeclarations(xmlStr)
	}

	return xmlStr
}

func (ne *NamespaceEncoder) addNamespaceDeclarations(xmlStr string) string {
	// Find first opening tag
	start := strings.Index(xmlStr, "<")
	if start == -1 {
		return xmlStr
	}

	end := strings.Index(xmlStr[start:], ">")
	if end == -1 {
		return xmlStr
	}
	end += start

	// Build namespace declarations
	var nsDecls strings.Builder
	for prefix, uri := range ne.namespaces {
		nsDecls.WriteString(" xmlns:")
		nsDecls.WriteString(prefix)
		nsDecls.WriteString(`="`)
		nsDecls.WriteString(uri)
		nsDecls.WriteString(`"`)
	}

	// Insert namespace declarations before closing >
	return xmlStr[:end] + nsDecls.String() + xmlStr[end:]
}

func (ne *NamespaceEncoder) addDefaultNamespacePrefix(xmlStr string) string {
	// Skip standard namespaces
	standardPrefixes := map[string]bool{
		"soap": true, "xsi": true, "xsd": true, "xml": true,
	}

	if standardPrefixes[ne.defaultNS] {
		return xmlStr
	}

	// Process entire XML string to add prefixes to all tags
	return ne.addPrefixToAllTags(xmlStr, ne.defaultNS)
}

func (ne *NamespaceEncoder) addPrefixToAllTags(xmlStr, prefix string) string {
	result := xmlStr
	offset := 0

	for {
		start := strings.Index(result[offset:], "<")
		if start == -1 {
			break
		}
		start += offset

		end := strings.Index(result[start:], ">")
		if end == -1 {
			break
		}
		end += start

		tag := result[start+1 : end]

		// Skip XML declarations, comments, and CDATA
		if strings.HasPrefix(tag, "?") || strings.HasPrefix(tag, "!") {
			offset = end + 1
			continue
		}

		// Handle closing tags
		if strings.HasPrefix(tag, "/") {
			tagName := tag[1:]
			if !strings.Contains(tagName, ":") {
				result = result[:start+1] + "/" + prefix + ":" + tagName + result[end:]
				end += len(prefix) + 1
			}
			offset = end + 1
			continue
		}

		// Handle opening tags (including self-closing)
		// Extract element name (before space or /)
		tagName := tag
		if spaceIdx := strings.Index(tag, " "); spaceIdx != -1 {
			tagName = tag[:spaceIdx]
		} else if strings.HasSuffix(tag, "/") {
			tagName = tag[:len(tag)-1]
		}

		// Only add prefix if element name doesn't already have one
		if !strings.Contains(tagName, ":") {
			if spaceIdx := strings.Index(tag, " "); spaceIdx != -1 {
				tag = prefix + ":" + tagName + tag[spaceIdx:]
				// Convert xmlns="..." to xmlns:prefix="..."
				tag = strings.Replace(tag, ` xmlns="`, ` xmlns:`+prefix+`="`, 1)
				// Add xmlns:xsi if element has xsi: attributes (insert at beginning)
				if strings.Contains(tag, "xsi:") && !strings.Contains(tag, "xmlns:xsi=") {
					if xsiURI, exists := ne.namespaces["xsi"]; exists {
						// Insert xmlns:xsi right after element name
						spacePos := strings.Index(tag, " ")
						if spacePos != -1 {
							tag = tag[:spacePos] + ` xmlns:xsi="` + xsiURI + `"` + tag[spacePos:]
						}
					}
				}
			} else if strings.HasSuffix(tag, "/") {
				tag = prefix + ":" + tagName + "/"
			} else {
				tag = prefix + ":" + tag
			}

			result = result[:start+1] + tag + result[end:]
			end += len(prefix) + 1
		} else {
			// Even if no prefix added, check for xsi: attributes
			if strings.Contains(tag, "xsi:") && !strings.Contains(tag, "xmlns:xsi=") {
				if xsiURI, exists := ne.namespaces["xsi"]; exists {
					spacePos := strings.Index(tag, " ")
					if spacePos != -1 {
						tag = tag[:spacePos] + ` xmlns:xsi="` + xsiURI + `"` + tag[spacePos:]
						result = result[:start+1] + tag + result[end:]
						end += len(` xmlns:xsi="` + xsiURI + `"`)
					}
				}
			}
		}

		offset = end + 1
	}

	return result
}

func (ne *NamespaceEncoder) addXsiNamespaceDeclarations(xmlStr string) string {
	result := xmlStr
	offset := 0

	for {
		start := strings.Index(result[offset:], "<")
		if start == -1 {
			break
		}
		start += offset

		end := strings.Index(result[start:], ">")
		if end == -1 {
			break
		}
		end += start

		tag := result[start+1 : end]

		// Skip XML declarations, comments, and closing tags
		if strings.HasPrefix(tag, "?") || strings.HasPrefix(tag, "!") || strings.HasPrefix(tag, "/") {
			offset = end + 1
			continue
		}

		// Check for xsi: attributes
		if strings.Contains(tag, "xsi:") && !strings.Contains(tag, "xmlns:xsi=") {
			if xsiURI, exists := ne.namespaces["xsi"]; exists {
				spacePos := strings.Index(tag, " ")
				if spacePos != -1 {
					tag = tag[:spacePos] + ` xmlns:xsi="` + xsiURI + `"` + tag[spacePos:]
					result = result[:start+1] + tag + result[end:]
					end += len(` xmlns:xsi="` + xsiURI + `"`)
				}
			}
		}

		offset = end + 1
	}

	return result
}

func (ne *NamespaceEncoder) GetResult() string {
	return ne.buffer.String()
}
