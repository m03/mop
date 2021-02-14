// Copyright (c) 2013-2019 by Michael Dvorkin and contributors. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/mop-tracker/mop/internal/config"
)

// const quotesURL = `http://download.finance.yahoo.com/d/quotes.csv?s=%s&f=sl1c1p2oghjkva2r2rdyj3j1`
const quotesURLv7 = `https://query1.finance.yahoo.com/v7/finance/quote?symbols=%s`
const quotesURLv7QueryParts = `&range=1d&interval=5m&indicators=close&includeTimestamps=false&includePrePost=false&corsDomain=finance.yahoo.com&.tsrc=finance`

const noDataIndicator = `N/A`

// Stock stores quote information for the particular stock ticker. The data
// for all the fields except 'Advancing' is fetched using Yahoo market API.
type Stock struct {
	Ticker     string `json:"symbol"`                      // Stock ticker.
	LastTrade  string `json:"regularMarketPrice"`          // l1: last trade.
	Change     string `json:"regularMarketChange"`         // c6: change real time.
	ChangePct  string `json:"regularMarketChangePercent"`  // k2: percent change real time.
	Open       string `json:"regularMarketOpen"`           // o: market open price.
	Low        string `json:"regularMarketDayLow"`         // g: day's low.
	High       string `json:"regularMarketDayHigh"`        // h: day's high.
	Low52      string `json:"fiftyTwoWeekLow"`             // j: 52-weeks low.
	High52     string `json:"fiftyTwoWeekHigh"`            // k: 52-weeks high.
	Volume     string `json:"regularMarketVolume"`         // v: volume.
	AvgVolume  string `json:"averageDailyVolume10Day"`     // a2: average volume.
	PeRatio    string `json:"trailingPE"`                  // r2: P/E ration real time.
	PeRatioX   string `json:"trailingPE"`                  // r: P/E ration (fallback when real time is N/A).
	Dividend   string `json:"trailingAnnualDividendRate"`  // d: dividend.
	Yield      string `json:"trailingAnnualDividendYield"` // y: dividend yield.
	MarketCap  string `json:"marketCap"`                   // j3: market cap real time.
	MarketCapX string `json:"marketCap"`                   // j1: market cap (fallback when real time is N/A).
	Currency   string `json:"currency"`                    // String code for currency of stock.
	Advancing  bool   // True when change is >= $0.
	PreOpen    string `json:"preMarketChangePercent,omitempty"`
	AfterHours string `json:"postMarketChangePercent,omitempty"`
}

// Quotes stores relevant pointers as well as the array of stock quotes for
// the tickers we are tracking.
type Quotes struct {
	market  *Market  // Pointer to Market.
	Profile *config.Profile // Pointer to Profile.
	Stocks  []Stock  // Array of stock quote data.
	errors  string   // Error string if any.
}

// Sets the initial values and returns new Quotes struct.
func NewQuotes(market *Market, profile *config.Profile) *Quotes {
	return &Quotes{
		market:  market,
		Profile: profile,
		errors:  ``,
	}
}

// Fetch the latest stock quotes and parse raw fetched data into array of
// []Stock structs.
func (quotes *Quotes) Fetch() (self *Quotes) {
	self = quotes // <-- This ensures we return correct quotes after recover() from panic().
	if quotes.isReady() {
		defer func() {
			if err := recover(); err != nil {
				quotes.errors = fmt.Sprintf("\n\n\n\nError fetching stock quotes...\n%s", err)
			} else {
				quotes.errors = ""
			}
		}()

		url := fmt.Sprintf(quotesURLv7, strings.Join(quotes.Profile.Tickers, `,`))
		response, err := http.Get(url + quotesURLv7QueryParts)
		if err != nil {
			panic(err)
		}

		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}

		quotes.parse2(body)
	}

	return quotes
}

// Ok returns two values: 1) boolean indicating whether the error has occured,
// and 2) the error text itself.
func (quotes *Quotes) Ok() (bool, string) {
	return quotes.errors == ``, quotes.errors
}

// AddTickers saves the list of tickers and refreshes the stock data if new
// tickers have been added. The function gets called from the line editor
// when user adds new stock tickers.
func (quotes *Quotes) AddTickers(tickers []string) (added int, err error) {
	if added, err = quotes.Profile.AddTickers(tickers); err == nil && added > 0 {
		quotes.Stocks = nil // Force fetch.
	}
	return
}

// RemoveTickers saves the list of tickers and refreshes the stock data if some
// tickers have been removed. The function gets called from the line editor
// when user removes existing stock tickers.
func (quotes *Quotes) RemoveTickers(tickers []string) (removed int, err error) {
	if removed, err = quotes.Profile.RemoveTickers(tickers); err == nil && removed > 0 {
		quotes.Stocks = nil // Force fetch.
	}
	return
}

// isReady returns true if we haven't fetched the quotes yet *or* the stock
// market is still open and we might want to grab the latest quotes. In both
// cases we make sure the list of requested tickers is not empty.
func (quotes *Quotes) isReady() bool {
	return (quotes.Stocks == nil || !quotes.market.IsClosed) && len(quotes.Profile.Tickers) > 0
}

// this will parse the json objects
func (quotes *Quotes) parse2(body []byte) (*Quotes, error) {
	// response -> quoteResponse -> result|error (array) -> map[string]interface{}
	// Stocks has non-int things
	// d := map[string]map[string][]Stock{}
	// some of these are numbers vs strings
	// d := map[string]map[string][]map[string]string{}
	d := map[string]map[string][]map[string]interface{}{}
	err := json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}
	results := d["quoteResponse"]["result"]

	quotes.Stocks = make([]Stock, len(results))
	for i, raw := range results {
		result := map[string]string{}
		for k, v := range raw {
			switch v.(type) {
			case string:
				result[k] = v.(string)
			case float64:
				result[k] = float2Str(v.(float64))
			default:
				result[k] = fmt.Sprintf("%v", v)
			}

		}
		quotes.Stocks[i].Ticker = result["symbol"]
		quotes.Stocks[i].LastTrade = result["regularMarketPrice"]
		quotes.Stocks[i].Change = result["regularMarketChange"]
		quotes.Stocks[i].ChangePct = result["regularMarketChangePercent"]
		quotes.Stocks[i].Open = result["regularMarketOpen"]
		quotes.Stocks[i].Low = result["regularMarketDayLow"]
		quotes.Stocks[i].High = result["regularMarketDayHigh"]
		quotes.Stocks[i].Low52 = result["fiftyTwoWeekLow"]
		quotes.Stocks[i].High52 = result["fiftyTwoWeekHigh"]
		quotes.Stocks[i].Volume = result["regularMarketVolume"]
		quotes.Stocks[i].AvgVolume = result["averageDailyVolume10Day"]
		quotes.Stocks[i].PeRatio = result["trailingPE"]
		// TODO calculate rt
		quotes.Stocks[i].PeRatioX = result["trailingPE"]
		quotes.Stocks[i].Dividend = result["trailingAnnualDividendRate"]
		quotes.Stocks[i].Yield = result["trailingAnnualDividendYield"]
		quotes.Stocks[i].MarketCap = result["marketCap"]
		// TODO calculate rt?
		quotes.Stocks[i].MarketCapX = result["marketCap"]
		quotes.Stocks[i].Currency = result["currency"]
		quotes.Stocks[i].PreOpen = result["preMarketChangePercent"]
		quotes.Stocks[i].AfterHours = result["postMarketChangePercent"]
		/*
			fmt.Println(i)
			fmt.Println("-------------------")
			for k, v := range result {
				fmt.Println(k, v)
			}
			fmt.Println("-------------------")
		*/
		adv, err := strconv.ParseFloat(quotes.Stocks[i].Change, 64)
		if err == nil {
			quotes.Stocks[i].Advancing = adv >= 0.0
		}
	}
	return quotes, nil
}

// Use reflection to parse and assign the quotes data fetched using the Yahoo
// market API.
func (quotes *Quotes) parse(body []byte) *Quotes {
	lines := bytes.Split(body, []byte{'\n'})
	quotes.Stocks = make([]Stock, len(lines))
	//
	// Get the total number of fields in the Stock struct. Skip the last
	// Advanicing field which is not fetched.
	//
	fieldsCount := reflect.ValueOf(quotes.Stocks[0]).NumField() - 1
	//
	// Split each line into columns, then iterate over the Stock struct
	// fields to assign column values.
	//
	for i, line := range lines {
		columns := bytes.Split(bytes.TrimSpace(line), []byte{','})
		for j := 0; j < fieldsCount; j++ {
			// ex. quotes.Stocks[i].Ticker = string(columns[0])
			reflect.ValueOf(&quotes.Stocks[i]).Elem().Field(j).SetString(string(columns[j]))
		}
		//
		// Try realtime value and revert to the last known if the
		// realtime is not available.
		//
		if quotes.Stocks[i].PeRatio == `N/A` && quotes.Stocks[i].PeRatioX != `N/A` {
			quotes.Stocks[i].PeRatio = quotes.Stocks[i].PeRatioX
		}
		if quotes.Stocks[i].MarketCap == `N/A` && quotes.Stocks[i].MarketCapX != `N/A` {
			quotes.Stocks[i].MarketCap = quotes.Stocks[i].MarketCapX
		}
		//
		// Stock is advancing if the change is not negative (i.e. $0.00
		// is also "advancing").
		//
		quotes.Stocks[i].Advancing = (quotes.Stocks[i].Change[0:1] != `-`)
	}

	return quotes
}

// -----------------------------------------------------------------------------
func sanitize(body []byte) []byte {
	return bytes.Replace(bytes.TrimSpace(body), []byte{'"'}, []byte{}, -1)
}

func float2Str(v float64) string {
	unit := ""
	switch {
	case v > 1.0e12:
		v = v / 1.0e12
		unit = "T"
	case v > 1.0e9:
		v = v / 1.0e9
		unit = "B"
	case v > 1.0e6:
		v = v / 1.0e6
		unit = "M"
	case v > 1.0e5:
		v = v / 1.0e3
		unit = "K"
	default:
		unit = ""
	}
	// parse
	return fmt.Sprintf("%0.3f%s", v, unit)
}
