package controller

import (
	"app/constants"
	"app/model"
	"app/provider"
	"app/webpojo"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

//LenderRateList get a list of lender rates based on the input criteria
func LenderRateList(w http.ResponseWriter, r *http.Request) {

	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		log.Println(readErr)
		ReturnError(w, readErr)
		return
	}
	log.Println("r.Body", string(body))

	lenderRateListReq := webpojo.LenderRateListReq{}
	jsonErr := json.Unmarshal(body, &lenderRateListReq)
	if jsonErr != nil {
		log.Println(jsonErr)
		ReturnError(w, jsonErr)
		return
	}

	var rates []model.LenderRate
	var dbErr error

	switch lenderRateListReq.Criteria {
	case "best":
		rates, dbErr = provider.GetBestLenderRatesList()
	default:
		rates, dbErr = model.LenderRatesListAll()
	}

	if dbErr != nil {
		log.Println(dbErr)
		ReturnError(w, dbErr)
		return
	}

	lenderRateListResp := new(webpojo.LenderRateListResp)
	var convErr error
	for _, v := range rates {
		rate := new(webpojo.LenderRatePojo)
		rate.Apr, convErr = strconv.ParseFloat(v.Apr, 64)
		if convErr != nil {
			log.Println(convErr)
			ReturnError(w, convErr)
			return
		}
		rate.BeginDate = v.BeginDate
		rate.EndDate = v.EndDate
		rate.ID = v.ID
		rate.Interest, convErr = strconv.ParseFloat(v.Interest, 64)
		if convErr != nil {
			log.Println(convErr)
			ReturnError(w, convErr)
			return
		}
		rate.LenderID = v.LenderID
		rate.LenderName = v.LenderName
		rate.Product = v.Product
		lenderRateListResp.LenderRates = append(lenderRateListResp.LenderRates, *rate)
	}

	lenderRateListResp.StatusCode = constants.StatusCode_200
	lenderRateListResp.Message = constants.Msg_200
	ReturnJsonResp(w, lenderRateListResp)
}

// FastQuotePost return info about best 30yrs rates for now
func FastQuotePost(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("error while get fast quote: can't read request body: ", err)
		ReturnCodeError(w, errors.New("error while read body"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	log.Println("fast quote request body: ", string(body))

	fastQuoteReq := &webpojo.FastQuoteReq{}
	err = json.Unmarshal(body, &fastQuoteReq)
	if err != nil {
		log.Println("error while get fast quote: error while unmarshal request body: ", err)
		ReturnCodeError(w, errors.New("can't unmarshal request body"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	fastQuote, err := provider.GetFastQuote(fastQuoteReq)
	if err != nil {
		log.Println("error while get fast quote: ", err)
		ReturnCodeError(w, errors.New("intenrnal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	ReturnCodeJSONResp(w, fastQuote, http.StatusOK)
}
