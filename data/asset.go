package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type yahooStock struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol   string `json:"symbol"`
				Timezone string `json:"timezone"`
			}
			TimeStamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int     `json:"volume"`
					High   []float64 `json:"high"`
					Open   []float64 `json:"open"`
				}
			}
		}
	}
}

type Metric struct {
	T     time.Time
	Value float64
}

type Price struct {
	TimeStamp time.Time
	Low       float64
	Close     float64
	Volume    int
	High      float64
	Open      float64
}

type Stock struct {
	Symbol    string
	TimeZone  string
	Indicator []Price
}

func newYahooStock() *yahooStock {
	return &yahooStock{}
}

/*dsour is yahoo,...*/
func NewStock(dsour string, symbol string, interv string, stime time.Time, etime time.Time) (*Stock, error) {
	/*TODO*/
	switch dsour {
	case "yahoo":
		ys := newYahooStock()
		ys.getData("testurl", symbol, interv, stime, etime)
		return converYahooStockToStock(ys)
	default:
		return nil, errors.New("Current datasource is not support...")

	}
}

func (ys *yahooStock) getData(testurl string, symbol string, interv string, stime time.Time, etime time.Time) {
	//fullUrl := fmt.Sprintf(yahooFinUrl, symbol, interv, stime.Unix(), etime.UTC())

	fullUrl := fmt.Sprintf(testurl, symbol, interv, strconv.FormatInt(stime.Unix(), 10), strconv.FormatInt(etime.Unix(), 10)) //for test

	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		slog.Info(fmt.Sprintf("New a http request error: %s", err.Error()))
		return
	}

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Info(fmt.Sprintf("GET url %s error: %s", fullUrl, err.Error()))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Info(fmt.Sprintf("Read url %s  body error: %s", fullUrl, err.Error()))
		return
	}

	err = json.Unmarshal(body, ys)
	if err != nil {
		slog.Info(fmt.Sprintf("Decode %s  body error: %s", fullUrl, err.Error()))
		return
	}
}

func converYahooStockToStock(ys *yahooStock) (*Stock, error) {
	if len(ys.Chart.Result) == 0 {
		slog.Info("Convert Yahoostock data to Stock is fail, the Yahoostock is null")
		return nil, errors.New("Convert Yahoostock data to Stock is fail, the Yahoostock is null")
	}

	s := &Stock{}
	symbol := ys.Chart.Result[0].Meta.Symbol
	ts := ys.Chart.Result[0].TimeStamp
	quote := ys.Chart.Result[0].Indicators.Quote[0]

	loc := time.FixedZone(ys.Chart.Result[0].Meta.Timezone, -4*3600) //force timezone record yahoostock

	for i, t := range ts {
		var p Price
		p.TimeStamp = time.Unix(t, 0).In(loc)
		p.Low = quote.Low[i]
		p.Close = quote.Close[i]
		p.Volume = quote.Volume[i]
		p.High = quote.High[i]
		p.Open = quote.Open[i]

		s.Indicator = append(s.Indicator, p)
	}
	s.Symbol = symbol
	return s, nil
}

/*
calculate RSV, the RSV formula is
n_day RSV = \frac{ClosingP_n - LowP_n}{HighP_n - LowP_n} * 100

period(n) must equal to len(prices)
*/
func RSV(period int, prices []Price) (float64, error) {
	if len(prices) == 0 {
		return 0, errors.New("Input prices is null.")
	}

	if period != len(prices) {
		return 0, errors.New("Period must equal to prices len.")
	}

	//closing price
	cp := prices[len(prices)-1].Close

	//find lowest and highest price during the period
	lp := prices[0].Low
	hp := prices[0].High
	for _, p := range prices {
		if p.Low < lp {
			lp = p.Low
		}

		if p.High > hp {
			hp = p.High
		}
	}

	if hp == lp {
		return 50, nil
	}

	return (cp - lp) / (hp - lp) * 100, nil

}

func (s *Stock) GetKDJMetrics(period int) ([]Metric, []Metric, []Metric, error) {
	if period > len(s.Indicator) {
		return nil, nil, nil, errors.New("Period should great than len of stock indicators")
	}

	k := make([]Metric, 0)
	d := make([]Metric, 0)
	j := make([]Metric, 0)

	//init kmetric
	kMetr := 50.0
	//init dmetric
	dMetr := 50.0

	//init jmetric
	jMetr := 0.0

	//start of window
	swindow := 0
	//end of window
	ewindow := period

	for ewindow <= len(s.Indicator) {
		window := s.Indicator[swindow:ewindow]
		rsv, err := RSV(period, window)
		if err != nil {
			return nil, nil, nil, err
		}

		kMetr = (2.0/3.0)*kMetr + (1.0/3.0)*rsv
		dMetr = (2.0/3.0)*dMetr + (1.0/3.0)*kMetr
		jMetr = 3*kMetr - 2*dMetr
		t := window[len(window)-1].TimeStamp

		k = append(k, Metric{T: t, Value: kMetr})
		d = append(d, Metric{T: t, Value: dMetr})
		j = append(j, Metric{T: t, Value: jMetr})

		//move window
		swindow++
		ewindow++

	}

	return k, d, j, nil

}

func (s *Stock) SMA(period int) ([]Metric, error) {
	SMAMetr := make([]Metric, 0)
	//TODO
	return SMAMetr, nil
}

func (s *Stock) EMA(period int) ([]Metric, error) {
	EMAMetr := make([]Metric, 0)
	//TODO
	return EMAMetr, nil
}

func (s *Stock) MACD(period int) ([]Metric, error) {
	MACDMetr := make([]Metric, 0)
	//TODO
	return MACDMetr, nil
}
