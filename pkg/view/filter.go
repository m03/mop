// Copyright (c) 2013-2019 by Michael Dvorkin and contributors. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package view

import (
	"strings"

	"github.com/mop-tracker/mop/internal/config"
	"github.com/mop-tracker/mop/pkg/model"
)

// Filter gets called to sort stock quotes by one of the columns. The
// setup is rather lengthy; there should probably be more concise way
// that uses reflection and avoids hardcoding the column names.
type Filter struct {
	profile *config.Profile // Pointer to where we store sort column and order.
}

// Returns new Filter struct.
func NewFilter(profile *config.Profile) *Filter {
	return &Filter{
		profile: profile,
	}
}

// Apply builds a list of sort interface based on current sort
// order, then calls sort.Sort to do the actual job.
func (filter *Filter) Apply(stocks []model.Stock) []model.Stock {
	var filteredStocks []model.Stock

	for _, stock := range stocks {
		var values = map[string]interface{}{
			"ticker":        strings.TrimSpace(stock.Ticker),
			"last":          m(stock.LastTrade),
			"change":        c(stock.Change),
			"changePercent": c(stock.ChangePct),
			"open":          m(stock.Open),
			"low":           m(stock.Low),
			"high":          m(stock.High),
			"low52":         m(stock.Low52),
			"high52":        m(stock.High52),
			"volume":        m(stock.Volume),
			"avgVolume":     m(stock.AvgVolume),
			"pe":            m(stock.PeRatio),
			"peX":           m(stock.PeRatioX),
			"dividend":      m(stock.Dividend),
			"yield":         m(stock.Yield),
			"mktCap":        m(stock.MarketCap),
			"mktCapX":       m(stock.MarketCapX),
			"advancing":     stock.Advancing,
		}

		result, err := filter.profile.FilterExpression.Evaluate(values)

		if err != nil {
			panic(err)
		}

		truthy, ok := result.(bool)

		if !ok {
			panic("Expression `" + filter.profile.Filter + "` should return a boolean value")
		}

		if truthy {
			filteredStocks = append(filteredStocks, stock)
		}
	}

	return filteredStocks
}
