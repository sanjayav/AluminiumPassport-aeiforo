package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"aluminium-passport/internal/models"
)

func SignVerifiableClaim(claim *models.VerifiableClaim, privateKey string) *models.VerifiableClaim {
	// Serialize the claim's subject to hash it (deterministic ordering)
	subject := claim.CredentialSubject
	keys := make([]string, 0, len(subject))
	for k := range subject {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	hash := sha256.New()
	for _, k := range keys {
		hash.Write([]byte(k))
		hash.Write([]byte(":"))
		hash.Write([]byte(fmt.Sprintf("%v", subject[k])))
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
