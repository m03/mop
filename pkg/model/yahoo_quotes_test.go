// Copyright (c) 2013-2016 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package model

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mop-tracker/mop/internal/config"
)

func TestQuotes(t *testing.T) {
	market := NewMarket()
	profile := config.NewProfile("../../moprc-sample")

	profile.Tickers = []string{"GOOG", "BA"}

	quotes := NewQuotes(market, profile)
	require.NotNil(t, quotes)

	data, err := ioutil.ReadFile("../../test/data/yahoo_quotes_sample.json")
	require.Nil(t, err)
	require.NotNil(t, data)

	require.True(t, quotes.isReady())
	//quotes.Fetch(data)
	_, err = quotes.parse2(data)
	assert.NoError(t, err)

	require.Equal(t, 2, len(quotes.Stocks))
	assert.Equal(t, "BA", quotes.Stocks[0].Ticker)
	assert.Equal(t, "331.76", quotes.Stocks[0].LastTrade)
	assert.Equal(t, "GOOG", quotes.Stocks[1].Ticker)
	assert.Equal(t, "1214.38", quotes.Stocks[1].LastTrade)
}
