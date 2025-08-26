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

type Stock struct {
	Symbol    string
	TimeZone  string
	Indicator []struct {
		TimeStamp time.Time
		Low       float64
		Close     float64
		Volume    int
		High      float64
		Open      float64
	}
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
		var tmpIndicator struct {
			TimeStamp time.Time
			Low       float64
			Close     float64
			Volume    int
			High      float64
			Open      float64
		}
		tmpIndicator.TimeStamp = time.Unix(t, 0).In(loc)
		tmpIndicator.Low = quote.Low[i]
		tmpIndicator.Close = quote.Close[i]
		tmpIndicator.Volume = quote.Volume[i]
		tmpIndicator.High = quote.High[i]
		tmpIndicator.Open = quote.Open[i]

		s.Indicator = append(s.Indicator, tmpIndicator)
	}
	s.Symbol = symbol
	return s, nil
}
