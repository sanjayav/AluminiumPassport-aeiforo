
package services

import (
    "crypto/sha256"
    "encoding/hex"
    "time"

    "aluminium-passport/internal/models"
)

func SignVerifiableClaim(claim *models.VerifiableClaim, privateKey string) *models.VerifiableClaim {
    // Serialize the claim's subject to hash it
    subject := claim.CredentialSubject
    hash := sha256.New()

    for k, v := range subject {
        hash.Write([]byte(k))
        hash.Write([]byte(":"))
        hash.Write([]byte(fmt.Sprintf("%v", v)))
        hash.Write([]byte(";"))
    }

    signature := hex.EncodeToString(hash.Sum(nil))

    claim.Proof = models.Proof{
        Type:               "Sha256Signature2025",
        Created:            time.Now().Format(time.RFC3339),
        ProofPurpose:       "assertionMethod",
        VerificationMethod: "did:example:issuer#key-1",
        SignatureValue:     signature,
    }

    return claim
}
