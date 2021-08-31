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
	GetBalance(exchangeAccountId int) (BalanceResponse, error)
	GetPortfolios(exchangeAccountId int) ([]PortfolioResponse, error)
	GetTicker(exchangeName string) (TickerResponse, error)
	UpdatePortfolio(exchangeAccountId, portfolioId int, request PortfolioUpdateRequest) error
	ActivatePortfolio(exchangeAccountId, portfolioID int) error
	RebalanceAccount(exchangeAccountId int) error
}

// MustNewShrimpyPrediction initializes a shrimpy client
func MustNewShrimpy(baseURL, apiKey, apiSecret string, logger *zap.Logger) Shrimpy {
	return &shrimpy{
		baseURL:   baseURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		logger:    logger,
	}
}

// shrimpy client implementation
type shrimpy struct {
	baseURL   string
	apiKey    string
	apiSecret string
	logger    *zap.Logger
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

// GetBalance retrieves balance data
func (s *shrimpy) GetBalance(exchangeAccountId int) (ret BalanceResponse, err error) {
	// prepare the base request
	url := fmt.Sprintf("%s/v1/accounts/%d/balance", s.baseURL, exchangeAccountId)
	s.logger.Debug("retrieving shrimpy account balance", zap.String("url", url))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		s.logger.Error("error generating request", zap.Error(err))
		return
	}
	resp, err := s.doRequest(req, http.StatusOK, "")
	if err != nil {
		s.logger.Error("request error", zap.Error(err))
		return
	}

	// handle response
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if err = json.Unmarshal(body, &ret); err != nil {
		s.logger.Error("unable to parse response", zap.Error(err), zap.String("body", string(body)))
		return
	}

	s.logger.Debug("successfully retrieved account balance", zap.Any("balance", ret))
	return
}

// GetPortfolios retreives a list of account automations
func (s *shrimpy) GetPortfolios(exchangeAccountId int) (ret []PortfolioResponse, err error) {
	// prepare the base request
	url := fmt.Sprintf("%s/v1/accounts/%d/portfolios", s.baseURL, exchangeAccountId)
	s.logger.Debug("retrieving shrimpy account portfolios", zap.String("url", url))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		s.logger.Error("error generating request", zap.Error(err))
		return
	}
	resp, err := s.doRequest(req, http.StatusOK, "")
	if err != nil {
		s.logger.Error("request error", zap.Error(err))
		return
	}

	// handle response
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if err = json.Unmarshal(body, &ret); err != nil {
		s.logger.Error("unable to parse response", zap.Error(err), zap.String("body", string(body)))
		return
	}

	s.logger.Debug("successfully retrieved portfolios", zap.String("body", string(body)), zap.Any("portfolios", ret))
	return
}

// GetTicker retrieves current exchange prices
func (s *shrimpy) GetTicker(exchangeName string) (ret TickerResponse, err error) {
	// prepare the base request
	url := fmt.Sprintf("%s/v1/%s/ticker", s.baseURL, strings.ToLower(exchangeName))
	s.logger.Debug("retrieving exchange ticker", zap.String("url", url))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		s.logger.Error("error generating request", zap.Error(err))
		return
	}
	resp, err := s.doRequest(req, http.StatusOK, "")
	if err != nil {
		s.logger.Error("request error", zap.Error(err))
		return
	}

	// handle response
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if err = json.Unmarshal(body, &ret); err != nil {
		s.logger.Error("unable to parse response", zap.Error(err), zap.String("body", string(body)))
		return
	}

	s.logger.Debug("successfully retrieved exchange ticker", zap.String("body", string(body)), zap.Any("data", ret))
	return
}

// UpdatePortfolio updates a specified account automation
func (s *shrimpy) UpdatePortfolio(exchangeAccountId, portfolioId int, request PortfolioUpdateRequest) (err error) {
	// prepare the base request
	url := fmt.Sprintf("%s/v1/accounts/%d/portfolios/%d/update", s.baseURL, exchangeAccountId, portfolioId)
	s.logger.Debug("updating portfolio", zap.String("url", url))
	requestBody, err := json.Marshal(request)
	if err != nil {
		s.logger.Error("unable to marshal JSON", zap.Error(err), zap.Any("request", request))
	}
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(requestBody)))
	if err != nil {
		s.logger.Error("error generating request", zap.Error(err))
		return
	}
	resp, err := s.doRequest(req, http.StatusOK, string(requestBody))
	if err != nil {
		s.logger.Error("request error", zap.Error(err))
		return
	}

	// handle response
	var ret RebalanceResponse
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if err = json.Unmarshal(body, &ret); err != nil {
		s.logger.Error("unable to parse response", zap.Error(err), zap.String("body", string(body)))
		return
	}
	if !ret.Success {
		s.logger.Debug("portfolio could not be updated")
		err = errors.New("unable to update portfolio")
		return
	}

	s.logger.Debug("successfully updated portfolio")
	return
}

// RebalanceAccount request rebalance on the default portfolio
func (s *shrimpy) RebalanceAccount(exchangeAccountId int) (err error) {
	// prepare the base request
	url := fmt.Sprintf("%s/v1/accounts/%d/rebalance", s.baseURL, exchangeAccountId)
	s.logger.Debug("rebalancing", zap.String("url", url))
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		s.logger.Error("error generating request", zap.Error(err))
		return
	}
	resp, err := s.doRequest(req, http.StatusOK, "")
	if err != nil {
		s.logger.Error("request error", zap.Error(err))
		return
	}

	// handle response
	var ret RebalanceResponse
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if err = json.Unmarshal(body, &ret); err != nil {
		s.logger.Error("unable to parse response", zap.Error(err), zap.String("body", string(body)))
		return
	}
	if !ret.Success {
		s.logger.Debug("portfolio could not be rebalanced")
		err = errors.New("unable to rebalance portfolio")
		return
	}

	s.logger.Debug("successfully rebalanced portfolio")
	return
}

// ActivatePortfolio enables or disables a specified automation
func (s *shrimpy) ActivatePortfolio(exchangeAccountId, portfolioID int) (err error) {
	// prepare the base request
	url := fmt.Sprintf("%s/v1/accounts/%d/portfolios/%d/activate", s.baseURL, exchangeAccountId, portfolioID)
	s.logger.Debug("activating shrimpy account portfolio", zap.String("url", url))
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		s.logger.Error("error generating request", zap.Error(err))
		return
	}
	resp, err := s.doRequest(req, http.StatusOK, "")
	if err != nil {
		s.logger.Error("request error", zap.Error(err))
		return
	}

	// handle response
	var ret ActivatePortfolioResponse
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if err = json.Unmarshal(body, &ret); err != nil {
		s.logger.Error("unable to parse response", zap.Error(err), zap.String("body", string(body)))
		return
	}
	if !ret.Success {
		s.logger.Debug("portfolio could not be activated")
		err = errors.New("unable to activate portfolio")
		return
	}

	s.logger.Debug("successfully activated portfolio")
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
