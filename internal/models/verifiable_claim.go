
package models

type VerifiableClaim struct {
    Context           []string               `json:"@context"`
    Type              []string               `json:"type"`
    Issuer            string                 `json:"issuer"`
    IssuanceDate      string                 `json:"issuanceDate"`
    CredentialSubject map[string]interface{} `json:"credentialSubject"`
    Proof             Proof                  `json:"proof"`
}

type Proof struct {
    Type               string `json:"type"`
    Created            string `json:"created"`
    ProofPurpose       string `json:"proofPurpose"`
    VerificationMethod string `json:"verificationMethod"`
    SignatureValue     string `json:"signatureValue"`
}
