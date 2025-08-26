package data

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestYahooStock_getdata(t *testing.T) {
	/*

	 */
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(yahooMockResponse))
	}))
	defer s.Close()

	ys := newYahooStock()

	/*convert unix time to time.Time*/
	unixToTime := func(unixtime string) time.Time {
		i, _ := strconv.ParseInt(unixtime, 10, 64)
		return time.Unix(i, 0)
	}
	symbol := "AAAL"
	interval := "1d"
	sTimeUnix := unixToTime("1704067200")
	eTimeUnix := unixToTime("1706745600")

	mockurl := s.URL + "/%s?interval=%s&period1=%s&period2=%s"
	fmt.Println(mockurl)
	ys.getData(mockurl, symbol, interval, sTimeUnix, eTimeUnix)
	fmt.Println(ys)
	fmt.Println("--------------")
	stock, _ := converYahooStockToStock(ys)
	fmt.Println(stock)
}
