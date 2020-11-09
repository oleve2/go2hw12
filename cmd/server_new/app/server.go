package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/wool/go2hw11/pkg/card"
)

type Server struct {
	cardSvc *card.Service
	mux     *http.ServeMux
}

func NewServer(cardSvc *card.Service, mux *http.ServeMux) *Server {
	return &Server{cardSvc: cardSvc, mux: mux}
}

func (s *Server) Init() {
	s.mux.HandleFunc("/echo", s.handlerEcho)
	s.mux.HandleFunc("/purchaseCard", s.handlerPurchaseCard)
	s.mux.HandleFunc("/getusercards/", s.handlerGetUserCards)
}

// для Echo
var countryTz = map[string]string{
	"Moscow": "Europe/Moscow",
}

func timeIn(name string) time.Time {
	loc, err := time.LoadLocation(countryTz[name])
	if err != nil {
		panic(err)
	}
	return time.Now().In(loc)
}

// ----------------------------------------------------------------
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ----------------------------------------------------------------
func (s *Server) handlerEcho(w http.ResponseWriter, r *http.Request) {
	loc, _ := time.LoadLocation("Europe/Moscow")
	resp := "ECHO " + time.Now().In(loc).String()
	_, err := w.Write([]byte(resp))
	if err != nil {
		log.Println(err)
	}
}

type PurchaseCardParams struct {
	CardType   string `json:"card_type"`
	CardIssuer string `json:"card_issuer"`
	UserID     int64  `json:"user_id"`
}

// ----------------------------------------------------------------
func (s *Server) handlerPurchaseCard(w http.ResponseWriter, r *http.Request) {
	var qparams PurchaseCardParams
	err := json.NewDecoder(r.Body).Decode(&qparams)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("params=", qparams)

	//
	err = card.CheckCardTypeCardIssuer(qparams.CardType, qparams.CardIssuer)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	//
	err = card.CheckUserID(s.cardSvc.GetCards(), qparams.UserID)
	if err != nil {
		http.Error(w, fmt.Sprintf("user %v does not exist", qparams.UserID), 400)
		return
	}

	//
	mxid := card.GetMaxIDFromcards(s.cardSvc.GetCards())
	s.cardSvc.SetCards(card.AddParamCardToCardslice(s.cardSvc.GetCards(), qparams.CardType, qparams.CardIssuer, qparams.UserID, mxid))
}

// ----------------------------------------------------------------
type userCards struct {
	CardsLength int64
	Cards       []*card.Card
}

//
type gucError struct {
}

func (s *Server) handlerGetUserCards(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	userID2, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		http.Error(w, "userid not parsed to int64", 400)
		return
	}
	err = card.CheckUserID(s.cardSvc.GetCards(), userID2)
	if err != nil {
		http.Error(w, fmt.Sprintf("user %v does not exist", userID2), 400)
		return
	}
	crdsUser := card.ReturnCardsByUserID(userID2, s.cardSvc.GetCards())
	crdsUserStruct := &userCards{CardsLength: int64(len(crdsUser)), Cards: crdsUser}

	crdsUserStructJSON, err := json.Marshal(crdsUserStruct)
	if err != nil {
		http.Error(w, "500 Internal Server Error", 500)
		log.Println(err.Error())
		return
	}
	//
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(crdsUserStructJSON)
}
