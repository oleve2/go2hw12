package card

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Card struct {
	ID           int64
	Type         string
	BankName     string
	CardNumber   string
	CardDueDate  string
	Balance      int64
	UserID       int64
	IsVirtual    bool
	Transactions []*Transaction
}

type Transaction struct {
	XMLName  string `xml:"transaction"`              //
	ID       int64  `json:"id" xml:"id"`             //
	TranType string `json:"trantype" xml:"trantype"` //
	TranSum  int64  `json:"transum" xml:"transum"`   //
	TranDate int64  `json:"trandate" xml:"trandate"` //  unix timestamp
	MccCode  string `json:"mcccode" xml:"mcccode"`   //
	Status   string `json:"status" xml:"status"`     //
	OwnerID  int64  `json:"ownerid" xml:"ownerid"`   //
}

type Transactions struct {
	XMLName      string         `xml:"transactions"`
	Transactions []*Transaction `xml:"transaction"`
}

type Service struct {
	//BankName string
	mu    sync.RWMutex
	Cards []*Card
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) SearchByNumber(number string) (*Card, bool) {
	for _, card := range s.Cards {
		if card.CardNumber == number {
			return card, true
		}
	}
	return nil, false
}

func AddTransaction(card *Card, transaction *Transaction) {
	card.Transactions = append(card.Transactions, transaction)
}

func valInSlice(val string, arr []string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

// SumByMCC - функция сумм по коду mmc
func SumByMCC(transactions []*Transaction, mcc []string) int64 {
	var totalMcc int64
	for _, v := range transactions {
		if valInSlice(v.MccCode, mcc) == true {
			totalMcc += v.TranSum
		}
	}
	return totalMcc
}

func PrintCardTrans(c *Card) {
	for _, v := range c.Transactions {
		fmt.Println(v.ID, v.TranDate, v.TranSum, v.TranType, v.MccCode, v.Status)
	}
}

func SortSlice(c *Card, asc bool) {
	if asc == true {
		sort.SliceStable(c.Transactions, func(i, j int) bool { return c.Transactions[i].TranSum < c.Transactions[j].TranSum })
	} else {
		sort.SliceStable(c.Transactions, func(i, j int) bool { return c.Transactions[i].TranSum > c.Transactions[j].TranSum })
	}
}

func Sum(transactions []int64) int64 {
	var res int64 = 0
	for _, v := range transactions {
		res += v
	}
	return res
}

func MakeTransMap(trans []*Transaction) map[string][]int64 {
	var mp = make(map[string][]int64)
	for _, v := range trans {
		var y string = strconv.Itoa(time.Unix(v.TranDate, 0).UTC().Year())
		var m string = strconv.FormatInt(int64(time.Unix(v.TranDate, 0).UTC().Month()), 10)

		var key string
		if len(m) == 1 {
			key = fmt.Sprintf("%s %s", y, m)
		} else if len(m) == 2 {
			key = fmt.Sprintf("%s 0%s", y, m)
		}
		mp[key] = append(mp[key], v.TranSum)
	}
	return mp
}

func SumConcurrently(trans []*Transaction, goroutines int) int64 {
	transMap := MakeTransMap(trans)

	lenTM := len(transMap)
	wg := sync.WaitGroup{}
	wg.Add(lenTM)

	total := int64(0)
	var sumByMonths = make(map[string]int64)
	mx := sync.Mutex{}

	for i, v := range transMap {
		yyyymm := i
		trans := v
		go func() {
			sum := Sum(trans)
			mx.Lock()
			sumByMonths[yyyymm] = sum
			total += sum
			mx.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()
	for k, v := range sumByMonths {
		fmt.Printf("%v : %d\n", k, v)
	}
	fmt.Println(sumByMonths)
	return total
}

/*
F1) Обычная функция, которая принимает на вход слайс транзакций и id владельца
- возвращает map с категориями и тратами по ним (сортировать они ничего не должна)

F2) Функция с mutex'ом, который защищает любые операции с map, соответственно, её задача: разделить слайс транзакций
на несколько кусков и в отдельных горутинах посчитать map'ы по кускам, после чего собрать всё в один большой map.
Важно: эта функция внутри себя должна вызывать функцию из п.1

F3) Функция с каналами, соответственно, её задача: разделить слайс транзакций на несколько кусков и в отдельных
горутинах посчитать map'ы по кускам, после чего собрать всё в один большой map (передавайте рассчитанные куски по каналу).
Важно: эта функция внутри себя должна вызывать функцию из п.1

F4) Функция с mutex'ом, который защищает любые операции с map, соответственно, её задача: разделить слайс транзакций
на несколько кусков и в отдельных горутинах посчитать, но теперь горутины напрямую пишут в общий map с результатами.
Важно: эта функция внутри себя не должна вызывать функцию из п.1

*/

// DiviveTranSlcToParts - разделить транзакции на NumberOfParts частей
func DiviveTranSlcToParts(tr []*Transaction, NumberOfParts int64) map[int64][]*Transaction {
	mp := make(map[int64][]*Transaction)
	slcLen := int64(len(tr))
	var partSize int64
	if slcLen%NumberOfParts == 0 {
		partSize = slcLen / NumberOfParts
	} else {
		partSize = slcLen/NumberOfParts + 1
	}

	var start, finish int64
	start = 0
	for i := 0; i < int(NumberOfParts); i++ {
		finish = start + partSize
		if finish < slcLen {
			mp[int64(i)] = tr[int(start):int(finish)]
		} else {
			mp[int64(i)] = tr[int(start):]
			return mp
		}
		start = finish
	}
	return mp
}

// F1 - сумма в лоб
func F1(tr []*Transaction, ownerID int64) map[string]int64 {
	mp := make(map[string]int64)
	for _, v := range tr {
		//fmt.Printf("code %v, code name %v, transum %d \n", v.MccCode, TranslateMCC(v.MccCode), v.TranSum)
		if v.OwnerID == ownerID {
			mp[TranslateMCC(v.MccCode)] += v.TranSum
		}
	}
	//fmt.Println(mp)
	return mp
}

// F2 - сумма конкурентно через мьютексы
func F2(tr []*Transaction, ownerID int64) map[string]int64 {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	result := make(map[string]int64)

	transSplit := DiviveTranSlcToParts(tr, 100)

	for _, v := range transSplit { // TODO здесь ваши условия разделения
		wg.Add(1)
		part := v
		go func() {
			m := F1(part, ownerID) // Categorize(part)
			mu.Lock()
			// TODO: вы перекладываете данные из m в result
			// TODO: подсказка - сделайте цикл по одной из map и смотрите, есть ли такие ключи в другой, если есть - прибавляйте
			for k, v := range m {
				result[k] += v
			}
			mu.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()
	return result
}

/*
Функция с каналами, соответственно, её задача: разделить слайс транзакций на несколько кусков и в отдельных горутинах посчитать map'ы по кускам,
после чего собрать всё в один большой map (передавайте рассчитанные куски по каналу).
Важно: эта функция внутри себя должна вызывать функцию из п.1
*/

// F3 - конкуретноый подсчет через каналы
func F3(tr []*Transaction, ownerID int64) map[string]int64 {
	result := make(map[string]int64)
	ch := make(chan map[string]int64)

	transSplit := DiviveTranSlcToParts(tr, 100)

	for _, v := range transSplit { // TODO здесь ваши условия разделения
		part := v // transactions[x:y]
		go func(ch chan<- map[string]int64) {
			ch <- F1(part, ownerID) //Categorize(part)
		}(ch)
	}

	partsCount := len(transSplit)
	finished := 0
	for value := range ch { // range result
		// TODO: вы перекладываете данные из m в result
		// TODO: подсказка - сделайте цикл по одной из map и смотрите, есть ли такие ключи в другой, если есть - прибавляйте
		for k, v := range value {
			result[k] += v
		}
		finished++
		if finished == partsCount {
			break
		}
	}
	return result
}

/*
F4) Функция с mutex'ом, который защищает любые операции с map, соответственно, её задача: разделить слайс транзакций
на несколько кусков и в отдельных горутинах посчитать, но теперь горутины напрямую пишут в общий map с результатами.
Важно: эта функция внутри себя не должна вызывать функцию из п.1
*/
// F4 -
func F4(tr []*Transaction, ownerID int64) map[string]int64 {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	result := make(map[string]int64)

	transSplit := DiviveTranSlcToParts(tr, 100)

	for _, v := range transSplit { // TODO здесь ваши условия разделения
		wg.Add(1)
		part := v //transactions[x:y]
		go func() {
			for _, t := range part {
				// TODO: 1. берём конкретную транзакцию
				// TODO: 2. смотрим, подходит ли по id владельца
				if t.OwnerID == ownerID {
					mu.Lock()
					// TODO: 3. если подходит, то закидываем в общий `map`
					result[TranslateMCC(t.MccCode)] += t.TranSum
					mu.Unlock()
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return result
}

/*
Экспорт и импорт транзакций (csv)

0) База
на вход - объект card с транзакцийми; путь экспорта + назв.файла;
на выход - ошибка (если её нет то nil)

1) Поля:
ID       int64
TranType string
TranSum  int64
TranDate int64 // unix timestamp
MccCode  string
Status   string
OwnerID  int64

2) План реализации:
) на вход - объект карт

*/

func MapRowToTransaction(s [][]string) []*Transaction {
	trans := make([]*Transaction, 0)
	for _, v := range s {
		id2, _ := strconv.ParseInt(v[0], 10, 64)
		transum2, _ := strconv.ParseInt(v[2], 10, 64)

		layout := "2006-01-02 15:04:05 +0300 MSK"
		trandate2, _ := time.Parse(layout, v[3]) //"2014-11-12T11:45:26.371Z"
		trandate3 := trandate2.Unix()

		owner2, _ := strconv.ParseInt(v[6], 10, 64)

		tr := &Transaction{
			"",
			id2,       //ID       int64
			v[1],      //TranType string
			transum2,  //TranSum  int64
			trandate3, //TranDate int64 // unix timestamp
			v[4],      //MccCode  string
			v[5],      //Status   string
			owner2,    //OwnerID  int64
		}
		trans = append(trans, tr)
	}
	return trans
}

func ExportToCSV(tr []*Transaction, exportPath string) error {
	if len(tr) == 0 {
		return nil
	}

	records := make([][]string, 0)
	for _, v := range tr {
		record := []string{
			strconv.FormatInt(v.ID, 10),
			v.TranType,
			strconv.FormatInt(v.TranSum, 10),
			time.Unix(v.TranDate, 0).String(), // TranDate
			v.MccCode,
			v.Status,
			strconv.FormatInt(v.OwnerID, 10),
		}
		records = append(records, record)
	}

	file, err := os.Create(exportPath)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func(c io.Closer) {
		if err := c.Close(); err != nil {
			log.Println(err)
		}
	}(file)

	w := csv.NewWriter(file)
	w.WriteAll(records)

	return nil
}

func ImportFromCSV(importPath string) ([]*Transaction, error) {
	file, err := os.Open(importPath)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func(c io.Closer) {
		if cerr := c.Close(); cerr != nil {
			log.Println(cerr)
		}
	}(file)

	reader := csv.NewReader(file)
	records := make([][]string, 0)
	for {
		record, err := reader.Read()
		if err != nil {
			if err != io.EOF {
				log.Println(err)
				return nil, err
			}
			//records = append(records, record)   <-- нужно ли прикреплять EOF к концу слайса? наверно нет ...
			break
		}
		records = append(records, record)
	}

	transConv := MapRowToTransaction(records)

	return transConv, nil
}

func ExporttoJSON(tr []*Transaction, exportPath string) error {
	if len(tr) == 0 {
		return nil
	}
	encData, err := json.Marshal(tr)
	if err != nil {
		log.Println(err)
		return err
	}
	err = ioutil.WriteFile(exportPath, encData, 0666)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func ImportFromJSON(importPath string) ([]*Transaction, error) {
	file, err := os.Open(importPath)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func(c io.Closer) {
		if cerr := c.Close(); cerr != nil {
			log.Println(cerr)
		}
	}(file)

	content, err := ioutil.ReadFile(importPath)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	var decoded []*Transaction
	err = json.Unmarshal(content, &decoded) // важно: передаём указатель
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//log.Println(reflect.TypeOf(decoded), decoded)
	return decoded, nil
}

func ExportXML(tr []*Transaction, exportPath string) error {
	if len(tr) == 0 {
		return nil
	}

	trs := &Transactions{Transactions: tr}
	encData, err := xml.Marshal(trs)
	if err != nil {
		log.Println(err)
		return err
	}
	encData = append([]byte(xml.Header), encData...)
	err = ioutil.WriteFile(exportPath, encData, 0666)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func ImportXML(importPath string) ([]*Transaction, error) {
	file, err := os.Open(importPath)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func(c io.Closer) {
		if cerr := c.Close(); cerr != nil {
			log.Println(cerr)
		}
	}(file)

	content, err := ioutil.ReadFile(importPath)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	//
	var decoded Transactions
	err = xml.Unmarshal(content, &decoded)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	log.Printf("%#v", decoded)
	trans := decoded.Transactions

	return trans, nil
}

func MakeCSV(tr []*Transaction) ([]byte, error) {
	if len(tr) == 0 {
		return nil, fmt.Errorf("transaction slice is empty")
	}
	buf := &bytes.Buffer{} // делать через буфер
	w := csv.NewWriter(buf)
	for _, v := range tr {
		record := []string{
			strconv.FormatInt(v.ID, 10),
			v.TranType,
			strconv.FormatInt(v.TranSum, 10),
			time.Unix(v.TranDate, 0).String(), // TranDate
			v.MccCode,
			v.Status,
			strconv.FormatInt(v.OwnerID, 10),
		}
		if err := w.Write(record); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), nil
}

func MakeJSON(tr []*Transaction) ([]byte, error) {
	if len(tr) == 0 {
		return nil, nil
	}
	encData, err := json.Marshal(tr)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return encData, nil
}

func MakeXML(tr []*Transaction) ([]byte, error) {
	if len(tr) == 0 {
		return nil, nil
	}
	trs := &Transactions{Transactions: tr}
	encData, err := xml.Marshal(trs)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	encData = append([]byte(xml.Header), encData...)
	return encData, nil
}

// InitCard - go2hw9 - для инициализации карты с транзакциями (для жкспорта из webapp)
func InitCard() *Card {
	card1 := &Card{ID: 1, Type: "Master", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01",
		Transactions: []*Transaction{
			&Transaction{ID: 1, TranType: "purchase", OwnerID: 2, TranSum: 1735_55, TranDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "5411", Status: "done Супермаркеты"},
			&Transaction{ID: 2, TranType: "purchase", OwnerID: 2, TranSum: 2000_00, TranDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "5411", Status: "done"},
			&Transaction{ID: 3, TranType: "purchase", OwnerID: 2, TranSum: 1203_91, TranDate: time.Date(2020, 2, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "5411", Status: "done Рестораны"},
			&Transaction{ID: 4, TranType: "purchase", OwnerID: 2, TranSum: 3562_21, TranDate: time.Date(2020, 2, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "1111", Status: ""},
			&Transaction{ID: 5, TranType: "purchase", OwnerID: 2, TranSum: 1111_11, TranDate: time.Date(2020, 3, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "1111", Status: ""},
			&Transaction{ID: 6, TranType: "purchase", OwnerID: 2, TranSum: 2222_22, TranDate: time.Date(2020, 3, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "1111", Status: ""},
			&Transaction{ID: 7, TranType: "purchase", OwnerID: 2, TranSum: 6666_66, TranDate: time.Date(2020, 4, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "3333", Status: ""},
			&Transaction{ID: 8, TranType: "purchase", OwnerID: 2, TranSum: 4444_44, TranDate: time.Date(2020, 4, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "3333", Status: ""},
			&Transaction{ID: 9, TranType: "purchase", OwnerID: 2, TranSum: 5555_55, TranDate: time.Date(2020, 5, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "5555", Status: ""},
			&Transaction{ID: 10, TranType: "purchase", OwnerID: 2, TranSum: 3333_33, TranDate: time.Date(2020, 5, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "5411", Status: ""},
			&Transaction{ID: 11, TranType: "purchase", OwnerID: 2, TranSum: 3333_33, TranDate: time.Date(2020, 5, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "5555", Status: ""},
			&Transaction{ID: 12, TranType: "purchase", OwnerID: 2, TranSum: 3333_33, TranDate: time.Date(2020, 5, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "5555", Status: ""},
			&Transaction{ID: 13, TranType: "purchase", OwnerID: 2, TranSum: 3333_33, TranDate: time.Date(2020, 5, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "5411", Status: ""},
		},
	}
	return card1
}

/*
==================================================================================
go2hw11 - функционал для сервера заказа пластиковых/виртуальных карт
*/

var (
	ErrCardFromBalanceLessThenAmount = errors.New("CardFrom balance < amount")
	ErrBothCardsNotFound             = errors.New("CardFrom and CardTo not found")
	ErrCardFromNotFound              = errors.New("CardFrom not found")
	ErrCardToNotFound                = errors.New("CardTo not found")
	ErrInvalidCardFromNumber         = errors.New("CardFrom number is not valid")
	ErrInvalidCardToNumber           = errors.New("CardTo number is not valid")

	ErrInvaildCardType   = errors.New("Card Type is not valid")
	ErrInvaildCardIssuer = errors.New("Card Issuer is not valid")
	ErrNoCardWithUserID  = errors.New("Cards with such UserID are not found")
)

func InitCardsHW11() []*Card {
	allCards := make([]*Card, 0)
	card11 := &Card{ID: 1, Type: "Master", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01", UserID: 1}
	card12 := &Card{ID: 2, Type: "Visa", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01", UserID: 1}

	card21 := &Card{ID: 3, Type: "Master", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01", UserID: 2}
	card22 := &Card{ID: 4, Type: "Visa", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01", UserID: 2}
	card23 := &Card{ID: 5, Type: "Master", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01", UserID: 2}

	card31 := &Card{ID: 6, Type: "Visa", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01", UserID: 3}
	card32 := &Card{ID: 7, Type: "Visa", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01", UserID: 3}
	card33 := &Card{ID: 8, Type: "Visa", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01", UserID: 3}

	card41 := &Card{ID: 9, Type: "UnionPay", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01", UserID: 4}

	allCards = append(allCards, card11, card12, card21, card22, card23, card31, card32, card33, card41)
	//
	return allCards
}

func (s *Service) ReturnCardsByUserID(userID int64, allCards []*Card) []*Card {
	// отбор карт по параметру ID пользователя
	userCards := make([]*Card, 0)
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, v := range allCards {
		if v.UserID == userID {
			userCards = append(userCards, v)
		}
	}
	return userCards
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

var CardTypes = []string{"plastic", "virtual"}

var CardIssuer = []string{"Master", "Visa", "UnionPay"}

func CheckCardTypeCardIssuer(ct string, ci string) error {
	if _, ok := Find(CardTypes, ct); !ok {
		return ErrInvaildCardType
	}
	if _, ok := Find(CardIssuer, ci); !ok {
		return ErrInvaildCardIssuer
	}
	return nil
}

func CheckUserID(crds []*Card, uid int64) error {
	var found bool
	found = false
	for _, v := range crds {
		//fmt.Println(i, v)
		if v.UserID == uid {
			found = true
		}
	}
	if found == true {
		return nil
	} else {
		return ErrNoCardWithUserID
	}
}

func GetMaxIDFromcards(crds []*Card) int64 {
	var newmxid int64 = crds[0].ID
	for _, v := range crds {
		if v.ID > newmxid {
			newmxid = v.ID
		}
	}
	// макс.значение +1
	return newmxid + 1
}

func (s *Service) AddParamCardToCardslice(crds []*Card, cardtype string, cardissuer string, userid int64, cardID int64) []*Card {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cardtype == "plastic" {
		c := &Card{
			ID: cardID, Type: cardissuer, BankName: "Tinkoff", CardNumber: "0000 0000 0000 0000",
			Balance: 0, CardDueDate: "2030-01-01", UserID: userid, IsVirtual: false,
		}
		crds = append(crds, c)
	}
	if cardtype == "virtual" {
		c := &Card{
			ID: cardID, Type: cardissuer, BankName: "Tinkoff", CardNumber: "0000 0000 0000 0000",
			Balance: 0, CardDueDate: "2030-01-01", UserID: userid, IsVirtual: true,
		}
		crds = append(crds, c)
	}
	return crds
}
