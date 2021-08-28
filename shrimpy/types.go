package shrimpy

import "time"

type AccountResponse struct {
	ID            int    `json:"id"`
	Exchange      string `json:"exchange"`
	Isrebalancing bool   `json:"isRebalancing"`
}

type BalanceResponse struct {
	Retrievedat time.Time `json:"retrievedAt"`
	Balances    []struct {
		Symbol      string  `json:"symbol"`
		Nativevalue float64 `json:"nativeValue"`
		Btcvalue    float64 `json:"btcValue"`
		Usdvalue    float64 `json:"usdValue"`
	} `json:"balances"`
}

type PortfolioUpdateRequest struct {
	Name               string                  `json:"name"`
	Rebalanceperiod    int                     `json:"rebalancePeriod"`
	Strategy           PortfolioUpdateStrategy `json:"strategy"`
	Strategytrigger    string                  `json:"strategyTrigger"`
	Rebalancethreshold string                  `json:"rebalanceThreshold"`
	Maxspread          string                  `json:"maxSpread"`
	Maxslippage        string                  `json:"maxSlippage"`
}

type PortfolioUpdateStrategy struct {
	Isdynamic   bool                        `json:"isDynamic"`
	Allocations []PortfolioUpdateAllocation `json:"allocations"`
}

type PortfolioUpdateAllocation struct {
	Symbol  string `json:"symbol"`
	Percent string `json:"percent"`
}
type PortfolioResponse struct {
	ID                 int                       `json:"id"`
	Name               string                    `json:"name"`
	Rebalanceperiod    int                       `json:"rebalancePeriod"`
	Active             bool                      `json:"active"`
	Strategy           PortfolioResponseStrategy `json:"strategy"`
	Strategytrigger    string                    `json:"strategyTrigger"`
	Rebalancethreshold string                    `json:"rebalanceThreshold"`
	Maxspread          string                    `json:"maxSpread"`
	Maxslippage        string                    `json:"maxSlippage"`
}
type PortfolioResponseStrategy struct {
	Isdynamic   bool                          `json:"isDynamic"`
	Allocations []PortfolioResponseAllocation `json:"allocations"`
}
type PortfolioResponseAllocation struct {
	Currency string `json:"currency"`
	Percent  string `json:"percent"`
	Fixed    bool   `json:"fixed"`
}

type ActivatePortfolioResponse struct {
	Success bool `json:"success"`
}

type RebalanceResponse struct {
	Success bool `json:"success"`
}

type AccountData struct {
	Name              string    `json:"name"`
	LastProcessedTime time.Time `json:"lastProcessed"`
	IsStale           bool      `json:"isStale"`
}
