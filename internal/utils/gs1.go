
package utils

import "fmt"

// GenerateGS1Link constructs a GS1 Digital Link using GTIN and serial number
func GenerateGS1Link(gtin, serial string) string {
    return fmt.Sprintf("https://id.gs1.org/01/%s/21/%s", gtin, serial)
}
