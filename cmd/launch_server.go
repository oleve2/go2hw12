package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/wool/go2hw11/pkg/card"
	//"github.com/wool/go2hw11/pkg/card"
)

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

func handlerEcho(w http.ResponseWriter, r *http.Request) {
	loc, _ := time.LoadLocation("Europe/Moscow")
	resp := "ECHO " + time.Now().In(loc).String()
	_, err := w.Write([]byte(resp))
	if err != nil {
		log.Println(err)
	}
}

/*
Шаг 1. Выбор услуги по заказу карты (можно выбрать как дополнительную, так и виртуальную):
Шаг 2. Заказ карты: (выберите счет и тип карты)

В это API должно быть передано:
1) Параметр тип карты (предложите сами в виде чего и какие допустимые значения)
2) Issuer (несмотря на то, что во фронтенде такого выбора нет, фронтенд может вам присылать этот
	скрытый от пользователя параметр, например, Visa или MasterCard)
3) Id пользователя

r.Context()  r.Body
*/

// PurchaseCardParams -
type PurchaseCardParams struct {
	CardType   string `json:"card_type"` // дополнительная или виртуальная
	CardIssuer string `json:"card_issuer"`
	UserID     int64  `json:"user_id"`
}

// HandlerPurchaseCard -
func handlerPurchaseCard(w http.ResponseWriter, r *http.Request) {
	// получение параметров из request.Body (передаются в json)
	var par1 PurchaseCardParams
	err := json.NewDecoder(r.Body).Decode(&par1)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("params=", par1)

	// инициализация списка всех доступных карт
	crds := card.InitCardsHW11()
	log.Println("cards inited")

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
	err = card.CheckUserID(crds, par1.UserID)
	if err != nil {
		//log.Println(err)
		http.Error(w, fmt.Sprintf("user %v does not exist", par1.UserID), 400)
		return
	}
	// если юзер есть - получение списка всех его карт
	crdsUser := card.ReturnCardsByUserID(par1.UserID, crds)

	// добавление новой карты в список карт
	mxid := card.GetMaxIDFromcards(crds)
	crdsUserNew := card.AddParamCardToCardslice(crdsUser, par1.CardType, par1.CardIssuer, par1.UserID, mxid)
	fmt.Println("updated card list:")
	for _, v := range crdsUserNew {
		fmt.Println(v)
	}
}

const defaultPort = "9999"
const defaultHost = "0.0.0.0"

// ==========================================
func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/echo", handlerEcho)
	mux.HandleFunc("/purchaseCard", handlerPurchaseCard)
	server := &http.Server{
		Addr:    "0.0.0.0:9999",
		Handler: mux,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
