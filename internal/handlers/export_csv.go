
package handlers

import (
    "encoding/csv"
    "net/http"
    "os"
    "strings"
)

// ExportAuditCSVHandler streams audit.log as a CSV file
func ExportAuditCSVHandler(w http.ResponseWriter, r *http.Request) {
    file, err := os.Open("audit.log")
    if err != nil {
        http.Error(w, "Failed to open audit log", http.StatusInternalServerError)
        return
    }
    defer file.Close()

    w.Header().Set("Content-Type", "text/csv")
    w.Header().Set("Content-Disposition", "attachment; filename=audit_log.csv")

    writer := csv.NewWriter(w)
    defer writer.Flush()

    writer.Write([]string{"Timestamp", "Username", "Role", "Action", "Resource"})

    var buf = make([]byte, 4096)
    n, _ := file.Read(buf)
    lines := strings.Split(string(buf[:n]), "\n")

    for _, line := range lines {
        if strings.TrimSpace(line) == "" {
            continue
        }
        // Example: [2025-07-31T12:45:10Z] USER=admin ROLE=issuer ACTION=CREATE RESOURCE=ALU123
        line = strings.TrimPrefix(line, "[")
        parts := strings.Split(line, "] ")
        if len(parts) < 2 {
            continue
        }
        timestamp := parts[0]
        fields := strings.Fields(parts[1])
        if len(fields) < 5 {
            continue
        }
        row := []string{timestamp, strings.Split(fields[0], "=")[1], strings.Split(fields[1], "=")[1], strings.Split(fields[2], "=")[1], strings.Split(fields[3], "=")[1]}
        writer.Write(row)
    }
}
