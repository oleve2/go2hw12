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

// Server -
type Server struct {
	cardSvc *card.Service
	mux     *http.ServeMux
}

// NewServer -
func NewServer(cardSvc *card.Service, mux *http.ServeMux) *Server {
	return &Server{cardSvc: cardSvc, mux: mux}
}

// Init -
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
// handlers
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ----------------------------------------------------------------
// handlerEcho
func (s *Server) handlerEcho(w http.ResponseWriter, r *http.Request) {
	loc, _ := time.LoadLocation("Europe/Moscow")
	resp := "ECHO " + time.Now().In(loc).String()
	_, err := w.Write([]byte(resp))
	if err != nil {
		log.Println(err)
	}
}

// PurchaseCardParams -
type PurchaseCardParams struct {
	CardType   string `json:"card_type"` // дополнительная или виртуальная
	CardIssuer string `json:"card_issuer"`
	UserID     int64  `json:"user_id"`
}

// ----------------------------------------------------------------
// HandlerPurchaseCard -
func (s *Server) handlerPurchaseCard(w http.ResponseWriter, r *http.Request) {
	// получение параметров из request.Body (передаются в json)
	var qparams PurchaseCardParams
	err := json.NewDecoder(r.Body).Decode(&qparams)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("params=", qparams)

	// инициализация списка карт в cardserver.execute
	// проверка card_type и card_issuer
	errCardType, errCardIssuer := card.CheckCardTypeCardIssuer(qparams.CardType, qparams.CardIssuer)
	if errCardType != nil {
		log.Println(errCardType)
		http.Error(w, "card type invalid", 400)
		return
	}
	if errCardIssuer != nil {
		log.Println(errCardIssuer)
		http.Error(w, "card issuer invalid", 400)
		return
	}

	// проверка UserID
	err = card.CheckUserID(s.cardSvc.Cards, qparams.UserID)
	if err != nil {
		http.Error(w, fmt.Sprintf("user %v does not exist", qparams.UserID), 400)
		return
	}

	// добавление новой карты в список карт
	mxid := card.GetMaxIDFromcards(s.cardSvc.Cards)
	s.cardSvc.Cards = card.AddParamCardToCardslice(s.cardSvc.Cards, qparams.CardType, qparams.CardIssuer, qparams.UserID, mxid)
	/*
		for _, v := range s.cardSvc.Cards {
			fmt.Println(v)
		}
	*/
}

// ----------------------------------------------------------------
// формат отдачи ответа
type userCards struct {
	CardsLength int64
	Cards       []*card.Card
}

// handlerGetUserCards -
func (s *Server) handlerGetUserCards(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	userID2, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		http.Error(w, "userid not parsed to int64", 400)
		return
	}
	err = card.CheckUserID(s.cardSvc.Cards, userID2)
	if err != nil {
		http.Error(w, fmt.Sprintf("user %v does not exist", userID2), 400)
		return
	}
	crdsUser := card.ReturnCardsByUserID(userID2, s.cardSvc.Cards)
	crdsUserStruct := &userCards{CardsLength: int64(len(crdsUser)), Cards: crdsUser}

	crdsUserStructJSON, err := json.Marshal(crdsUserStruct)
	if err != nil {
		http.Error(w, "error - userid cards not converted to json", 400)
		return
	}
	//
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte(crdsUserStructJSON))
}
