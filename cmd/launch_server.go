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
	}
	fmt.Println("params=", par1)

	// инициализация списка всех доступных карт
	crds := card.InitCardsHW11()
	// проверка card_type и card_issuer
	errCT, errCI := card.CheckCardTypeCardIssuer(par1.CardType, par1.CardIssuer)
	if errCT != nil {
		log.Println(errCT)
	}
	if errCI != nil {
		log.Println(errCI)
	}
	// проверка UserID
	fmt.Println("check UserID")
	err = card.CheckUserID(crds, par1.UserID)
	if err != nil {
		log.Println(err)
	}
	// добавление новой карты в список карт
	crds = card.AddParamCardToCardslice(crds, par1.CardType, par1.CardIssuer, par1.UserID)
	fmt.Println("updated card list:")
	for _, v := range crds {
		fmt.Println(v)
	}
}

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
