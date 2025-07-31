
package utils

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "os"
)

func UploadToIPFS(jsonData interface{}) (string, error) {
    // Marshal the input to JSON
    data, err := json.Marshal(jsonData)
    if err != nil {
        return "", err
    }

    // Prepare multipart form
    var b bytes.Buffer
    w := multipart.NewWriter(&b)
    fw, err := w.CreateFormFile("file", "passport.json")
    if err != nil {
        return "", err
    }
    if _, err = fw.Write(data); err != nil {
        return "", err
    }
    w.Close()

    // Read env
    projectID := os.Getenv("IPFS_PROJECT_ID")
    projectSecret := os.Getenv("IPFS_PROJECT_SECRET")
    apiURL := os.Getenv("IPFS_API_URL")

    if projectID == "" || projectSecret == "" || apiURL == "" {
        return "", fmt.Errorf("Missing IPFS config in environment")
    }

    // Build request
    req, err := http.NewRequest("POST", apiURL+"/api/v0/add", &b)
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", w.FormDataContentType())
    req.SetBasicAuth(projectID, projectSecret)

    // Send request
    client := &http.Client{}
    res, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer res.Body.Close()

    // Parse response
    body, _ := io.ReadAll(res.Body)
    var result map[string]interface{}
    if err := json.Unmarshal(body, &result); err != nil {
        return "", err
    }

    hash, ok := result["Hash"].(string)
    if !ok {
        return "", fmt.Errorf("Invalid IPFS response: %s", string(body))
    }

    return "ipfs://" + hash, nil
}
