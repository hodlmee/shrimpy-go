package shrimpy
 
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"
    "strconv"
    "strings"
    "time"
 
    "go.uber.org/zap"
)
 
// Shrimpy interface definition
type Shrimpy interface {
    GetAccounts() ([]AccountResponse, error)
}
 
// MustNewShrimpyPrediction gets a shrimpy prediction function
func MustNewShrimpy(url, apiKey, apiSecret string, logger *zap.Logger) Shrimpy {
    return &shrimpy{
        baseURL:   url,
        apiKey:    apiKey,
        apiSecret: apiSecret,
        logger:    logger,
    }
}
 
type shrimpy struct {
    baseURL   string
    apiKey    string
    apiSecret string
    logger    *zap.Logger
}
 
// AccountResponse the JSON payload returned from shrimpy API account list
type AccountResponse struct {
    ID            int    `json:"id"`
    Exchange      string `json:"exchange"`
    Isrebalancing bool   `json:"isRebalancing"`
}
 
// GetAccounts retrieves a list of exchange accounts managed by shrimpy
func (s *shrimpy) GetAccounts() (ret []AccountResponse, err error) {
 
    // prepare the base request
    url := fmt.Sprintf("%s/v1/accounts", s.baseURL)
    s.logger.Debug("retrieving shrimpy accounts", zap.String("url", url))
    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        s.logger.Error("error generating request", zap.Error(err))
        return nil, err
    }
    resp, err := s.doRequest(req, http.StatusOK, "")
    if err != nil {
        s.logger.Error("request error", zap.Error(err))
        return nil, err
    }
 
    // handle response
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    if err = json.Unmarshal(body, &ret); err != nil {
        s.logger.Error("unable to parse response", zap.Error(err), zap.String("body", string(body)))
        return
    }
    s.logger.Debug("successfully retrieved accounts", zap.Any("accounts", ret))
    return
}
 
// doRequest prepare and send request to shrimpy API
func (s *shrimpy) doRequest(req *http.Request, expectedRC int, body string) (*http.Response, error) {
    now := time.Now()
    nonce := strconv.FormatInt(now.Unix(), 10)
 
    // create signature
    signature, err := s.getSignature(req.URL.Path, req.Method, nonce, body)
    if err != nil {
        s.logger.Error("error generating signature", zap.Error(err))
        return nil, err
    }
 
    // add required headers
    req.Header.Add("Content-Type", "application/json")
    req.Header.Add("SHRIMPY-API-KEY", s.apiKey)
    req.Header.Add("SHRIMPY-API-NONCE", nonce)
    req.Header.Add("SHRIMPY-API-SIGNATURE", signature)
 
    // make the request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        s.logger.Error("error retrieving shrimpy request", zap.Error(err))
        return resp, err
    }
    if resp.StatusCode != expectedRC {
        defer resp.Body.Close()
        body, _ := ioutil.ReadAll(resp.Body)
        msg := fmt.Sprintf("unexpected response code %d: %s", resp.StatusCode, string(body))
        s.logger.Error(msg, zap.String("response", string(body)))
        return resp, errors.New(msg)
    }
    s.logger.Debug("request successful")
    return resp, nil
}
 
// getSignature generates a request signature
func (s *shrimpy) getSignature(path, method, nonce, body string) (string, error) {
    prehash := path + method + nonce + body
    key, err := base64.StdEncoding.DecodeString(s.apiSecret)
    if err != nil {
        s.logger.Error("unable to decode secret", zap.Error(err))
        return "", err
    }
    h := hmac.New(sha256.New, key)
    if _, err = h.Write([]byte(prehash)); err != nil {
        s.logger.Error("error writing prehash HMAC", zap.Error(err))
        return "", err
    }
    signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
    s.logger.Debug("successfully generated signature", zap.String("prehash", prehash), zap.String("signature", signature))
    return signature, nil
}
