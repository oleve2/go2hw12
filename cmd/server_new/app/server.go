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
	var par1 PurchaseCardParams
	err := json.NewDecoder(r.Body).Decode(&par1)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("params=", par1)

	// инициализация списка всех доступных карт
	//crds := s.cardSvc.Cards //card.InitCardsHW11()
	//log.Println("cards loaded from svc")

	// проверка card_type и card_issuer
	errCT, errCI := card.CheckCardTypeCardIssuer(par1.CardType, par1.CardIssuer)
	//log.Println(errCT, errCI)
	if errCT != nil {
		log.Println(errCT)
		http.Error(w, "card type invalid", 400)
		return
	}
	if errCI != nil {
		log.Println(errCI)
		http.Error(w, "card issuer invalid", 400)
		return
	}

	// проверка UserID
	fmt.Println("check UserID")
	err = card.CheckUserID(s.cardSvc.Cards, par1.UserID) // crds
	if err != nil {
		//log.Println(err)
		http.Error(w, fmt.Sprintf("user %v does not exist", par1.UserID), 400)
		return
	}
	// если юзер есть - получение списка всех его карт
	//crdsUser := card.ReturnCardsByUserID(par1.UserID, s.cardSvc.Cards)

	// добавление новой карты в список карт
	mxid := card.GetMaxIDFromcards(s.cardSvc.Cards)
	s.cardSvc.Cards = card.AddParamCardToCardslice(s.cardSvc.Cards, par1.CardType, par1.CardIssuer, par1.UserID, mxid)
	fmt.Println("updated card list:")
	for _, v := range s.cardSvc.Cards { // crdsUserNew
		fmt.Println(v)
	}
}

// ----------------------------------------------------------------
// handlerGetUserCards -
func (s *Server) handlerGetUserCards(w http.ResponseWriter, r *http.Request) {
	//crds := s.cardSvc.Cards //card.InitCardsHW11()
	// получение параметра из query
	userID := r.URL.Query()["userID"][0]
	userID2, err := strconv.ParseInt(userID, 10, 64)
	err = card.CheckUserID(s.cardSvc.Cards, userID2)
	if err != nil {
		//log.Println(err)
		http.Error(w, fmt.Sprintf("user %v does not exist", userID2), 400)
		return
	}
	crdsUser := card.ReturnCardsByUserID(userID2, s.cardSvc.Cards)
	fmt.Println(crdsUser)
	outCrdUsr, err := json.Marshal(crdsUser)
	//
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte(outCrdUsr))
}
