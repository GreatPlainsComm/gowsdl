package main

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/GreatPlainsComm/gowsdl/soap"
)

type ServiceInformationItem struct {
	XMLName   xml.Name `xml:"ServiceInformationItem"`
	XsiType   string   `xml:"xsi:type,attr"`
	CatalogID int      `xml:"CatalogID"`
}

func main() {
	// Test data
	item := ServiceInformationItem{
		XsiType:   "ServiceInformationItem",
		CatalogID: 114,
	}

	fmt.Printf("XsiType value: '%s'\n", item.XsiType)
	fmt.Printf("CatalogID value: %d\n", item.CatalogID)

	// Test 1: Standard XML encoding
	fmt.Println("\n=== Standard XML Encoding ===")
	xmlData, err := xml.MarshalIndent(item, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Standard XML:\n%s\n", string(xmlData))

	// Test 2: SOAP Client with NamespaceEncoder
	fmt.Println("\n=== SOAP Client with NamespaceEncoder ===")
	client := soap.NewClient("http://example.com")
	client.SetNamespaces(map[string]string{
		"ns1": "http://webservices.idibilling.com/OrderPlacement/1.0",
		"xsi": "http://www.w3.org/2001/XMLSchema-instance",
	})

	// Test with SOAP envelope (this will use NamespaceEncoder internally)
	envelope := struct {
		XMLName xml.Name `xml:"soap:Envelope"`
		XmlNS   string   `xml:"xmlns:soap,attr"`
		Body    struct {
			XMLName xml.Name                `xml:"soap:Body"`
			Content ServiceInformationItem `xml:",omitempty"`
		} `xml:"soap:Body"`
	}{
		XmlNS: "http://www.w3.org/2003/05/soap-envelope",
		Body: struct {
			XMLName xml.Name                `xml:"soap:Body"`
			Content ServiceInformationItem `xml:",omitempty"`
		}{
			Content: item,
		},
	}

	// This simulates what the SOAP client does internally
	xmlData2, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		panic(err)
	}
	result := string(xmlData2)
	fmt.Printf("NamespaceEncoder result:\n%s\n", result)

	// Check for expected attributes
	fmt.Println("\n=== Validation ===")
	fmt.Printf("Contains 'xsi:type': %t\n", strings.Contains(result, "xsi:type"))
	fmt.Printf("Contains 'xmlns:xsi': %t\n", strings.Contains(result, "xmlns:xsi"))
	fmt.Printf("Contains 'ns1:ServiceInformationItem': %t\n", strings.Contains(result, "ns1:ServiceInformationItem"))
}