package main

import (
	"bytes"
	"context"
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

type CreateCartRequest struct {
	XMLName xml.Name               `xml:"CreateCart"`
	Item    ServiceInformationItem `xml:"ServiceInformationItem"`
}

func main() {
	fmt.Println("=== Testing NamespaceEncoder with xsi:type ===\n")

	// Test 1: Standard XML encoding (baseline)
	fmt.Println("1. Standard XML encoding:")
	item := ServiceInformationItem{
		XsiType:   "ServiceInformationItem",
		CatalogID: 114,
	}

	xmlData, _ := xml.MarshalIndent(item, "", "  ")
	fmt.Printf("%s\n\n", string(xmlData))

	// Test 2: Direct NamespaceEncoder test
	fmt.Println("2. Direct NamespaceEncoder test:")

	buffer := &bytes.Buffer{}
	encoder := soap.NewNamespaceEncoder(buffer, map[string]string{
		"xsi": "http://www.w3.org/2001/XMLSchema-instance",
	}, "")

	err := encoder.Encode(item)
	if err != nil {
		fmt.Printf("Encode error: %v\n", err)
		return
	}

	err = encoder.Flush()
	if err != nil {
		fmt.Printf("Flush error: %v\n", err)
		return
	}

	result := buffer.String()
	fmt.Printf("Result: %s\n\n", result)

	// Test 3: SOAP client with namespaces
	fmt.Println("3. SOAP client with NamespaceEncoder:")

	client := soap.NewClient("http://example.com")
	// Don't set any custom namespaces - xsi should be added automatically

	request := CreateCartRequest{
		Item: item,
	}

	fmt.Println("Attempting SOAP call (will show debug output)...")
	err = client.CallContext(context.Background(), "CreateCart", request, nil)
	fmt.Printf("Call result: %v\n\n", err)

	// Test 4: Validation
	fmt.Println("4. Validation:")
	fmt.Printf("Contains 'xsi:type': %t\n", strings.Contains(result, "xsi:type"))
	fmt.Printf("Contains 'xmlns:xsi': %t\n", strings.Contains(result, "xmlns:xsi"))
	fmt.Printf("Contains 'ServiceInformationItem': %t\n", strings.Contains(result, "ServiceInformationItem"))

	if strings.Contains(result, "xmlns:xsi") && strings.Contains(result, "xsi:type") {
		fmt.Println("✅ SUCCESS: Both xmlns:xsi and xsi:type are present!")
	} else {
		fmt.Println("❌ FAILED: Missing xmlns:xsi or xsi:type")
	}
}
