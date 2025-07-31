
package handlers

import (
    "encoding/json"
    "net/http"
    "time"

    "aluminium-passport/internal/models"
    "aluminium-passport/internal/services"
)

func ExportSignedCredentialHandler(w http.ResponseWriter, r *http.Request) {
    claim := &models.VerifiableClaim{
        Context: []string{"https://www.w3.org/2018/credentials/v1"},
        Type:    []string{"VerifiableCredential", "AluminiumPassport"},
        Issuer:  "did:example:issuer",
        IssuanceDate: time.Now().Format(time.RFC3339),
        CredentialSubject: map[string]interface{}{
            "passportId": "ALU123",
            "batchId": "BATCH456",
            "bauxiteOrigin": "Australia",
            "carbonEmissionsPerKg": 6.2,
        },
    }

    signed := services.SignVerifiableClaim(claim, "dummy-private-key")

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Content-Disposition", "attachment; filename=verifiable-passport.json")
    json.NewEncoder(w).Encode(signed)
}
