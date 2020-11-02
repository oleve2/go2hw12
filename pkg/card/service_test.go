package card

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestMapRowToTransaction(t *testing.T) {
	trans := []*Transaction{
		&Transaction{ID: 1, TranType: "purchase", OwnerID: 2, TranSum: 1735_55, TranDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "5411", Status: "done Супермаркеты"},
		&Transaction{ID: 2, TranType: "purchase", OwnerID: 2, TranSum: 2000_00, TranDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "5411", Status: "done"},
		&Transaction{ID: 3, TranType: "purchase", OwnerID: 2, TranSum: 1203_91, TranDate: time.Date(2020, 2, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "5411", Status: "done Рестораны"},
	}
	transFromImport := [][]string{
		{"1", "purchase", "173555", "2020-01-01 00:00:00 +0300 MSK", "5411", "done", "Супермаркеты", "2"},
		{"2", "purchase", "200000", "2020-01-01 00:00:00 +0300 MSK", "5411", "done", "2"},
		{"3", "purchase", "120391", "2020-02-01 00:00:00 +0300 MSK", "5411", "done", "Рестораны", "2"},
	}
	fmt.Println(trans, transFromImport)

	type args struct {
		s [][]string
	}
	tests := []struct {
		name string
		args args
		want []*Transaction
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := MapRowToTransaction(tt.args.s); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("MapRowToTransaction() = %v, want %v", got, tt.want)
		}
	}
}

/*
// MakeTransactions - конструирование транзакций к карте по шаблону
func MakeTransactions() []*Transaction {
	const users = 10_000
	const transactionsPerUser = 10_000
	const transactionAmount = 1_00
	transactions := make([]*Transaction, users*transactionsPerUser)
	for index := range transactions {
		switch index % 100 {
		case 0:
			// Например, каждая 100-ая транзакция в банке от нашего юзера в категории такой-то (OwnerID = 2)
			transactions[index] = &Transaction{ID: int64(index), TranType: "purchase", OwnerID: 2, TranSum: transactionAmount, //  * int64(index)
				TranDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "5411", Status: "done"}
		case 20:
			// Например, каждая 120-ая транзакция в банке от нашего юзера в категории такой-то (OwnerID = 2)
			transactions[index] = &Transaction{ID: int64(index), TranType: "purchase", OwnerID: 2, TranSum: transactionAmount,
				TranDate: time.Date(2020, 5, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "5555", Status: "done"}
		default:
			// Транзакции других юзеров, нужны для "общей" массы
			transactions[index] = &Transaction{ID: int64(index), TranType: "purchase", OwnerID: int64(index), TranSum: transactionAmount,
				TranDate: time.Date(2020, 5, 1, 0, 0, 0, 0, time.Local).Unix(), MccCode: "3333", Status: "done"}
		}
	}
	return transactions
}

func TestF1(t *testing.T) {
	card1 := &Card{ID: 1, Type: "Master", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01"}
	card1.Transactions = MakeTransactions()

	wantMap := map[string]int64{"Категория 3333": 100, "Категория 5555": 100000000, "Супермаркеты": 100000000}

	type args struct {
		tr      []*Transaction
		ownerID int64
	}
	tests := []struct {
		name string
		args args
		want map[string]int64
	}{
		{
			name: "Test F1 with made trans",
			args: args{tr: card1.Transactions, ownerID: 2},
			want: wantMap,
		},
	}
	for _, tt := range tests {
		if got := F1(tt.args.tr, tt.args.ownerID); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("F1() = %v, want %v", got, tt.want)
		}
	}
}

func TestF2(t *testing.T) {
	card1 := &Card{ID: 1, Type: "Master", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01"}
	card1.Transactions = MakeTransactions()

	wantMap := map[string]int64{"Категория 3333": 100, "Категория 5555": 100000000, "Супермаркеты": 100000000}

	type args struct {
		tr      []*Transaction
		ownerID int64
	}
	tests := []struct {
		name string
		args args
		want map[string]int64
	}{
		{
			name: "Test F2 with made trans",
			args: args{tr: card1.Transactions, ownerID: 2},
			want: wantMap,
		},
	}
	for _, tt := range tests {
		if got := F2(tt.args.tr, tt.args.ownerID); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("F2() = %v, want %v", got, tt.want)
		}
	}
}

func TestF3(t *testing.T) {
	card1 := &Card{ID: 1, Type: "Master", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01"}
	card1.Transactions = MakeTransactions()

	wantMap := map[string]int64{"Категория 3333": 100, "Категория 5555": 100000000, "Супермаркеты": 100000000}

	type args struct {
		tr      []*Transaction
		ownerID int64
	}
	tests := []struct {
		name string
		args args
		want map[string]int64
	}{
		{
			name: "Test F3 with made trans",
			args: args{tr: card1.Transactions, ownerID: 2},
			want: wantMap,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := F3(tt.args.tr, tt.args.ownerID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("F3() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestF4(t *testing.T) {
	card1 := &Card{ID: 1, Type: "Master", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01"}
	card1.Transactions = MakeTransactions()

	wantMap := map[string]int64{"Категория 3333": 100, "Категория 5555": 100000000, "Супермаркеты": 100000000}

	type args struct {
		tr      []*Transaction
		ownerID int64
	}
	tests := []struct {
		name string
		args args
		want map[string]int64
	}{
		{
			name: "Test F4 with made trans",
			args: args{tr: card1.Transactions, ownerID: 2},
			want: wantMap,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := F4(tt.args.tr, tt.args.ownerID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("F4() = %v, want %v", got, tt.want)
			}
		})
	}
}

// go test -bench=. -benchtime=3x ./...
// benchmark F1
func BenchmarkF1(b *testing.B) {
	card1 := &Card{ID: 1, Type: "Master", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01"}
	card1.Transactions = MakeTransactions()

	want := map[string]int64{"Категория 3333": 100, "Категория 5555": 100000000, "Супермаркеты": 100000000}
	b.ResetTimer() // сбрасываем таймер, т.к. сама генерация транзакций достаточно ресурсоёмка
	for i := 0; i < b.N; i++ {
		result := F1(card1.Transactions, 2)
		b.StopTimer() // останавливаем таймер, чтобы время сравнения не учитывалось
		if !reflect.DeepEqual(result, want) {
			b.Fatalf("invalid result, got %v, want %v", result, want)
		}
		b.StartTimer() // продолжаем работу таймера
	}
}

// benchmark F2
func BenchmarkF2(b *testing.B) {
	card1 := &Card{ID: 1, Type: "Master", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01"}
	card1.Transactions = MakeTransactions()

	want := map[string]int64{"Категория 3333": 100, "Категория 5555": 100000000, "Супермаркеты": 100000000}
	b.ResetTimer() // сбрасываем таймер, т.к. сама генерация транзакций достаточно ресурсоёмка
	for i := 0; i < b.N; i++ {
		result := F2(card1.Transactions, 2)
		b.StopTimer() // останавливаем таймер, чтобы время сравнения не учитывалось
		if !reflect.DeepEqual(result, want) {
			b.Fatalf("invalid result, got %v, want %v", result, want)
		}
		b.StartTimer() // продолжаем работу таймера
	}
}

// benchmark F3
func BenchmarkF3(b *testing.B) {
	card1 := &Card{ID: 1, Type: "Master", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01"}
	card1.Transactions = MakeTransactions()

	want := map[string]int64{"Категория 3333": 100, "Категория 5555": 100000000, "Супермаркеты": 100000000}
	b.ResetTimer() // сбрасываем таймер, т.к. сама генерация транзакций достаточно ресурсоёмка
	for i := 0; i < b.N; i++ {
		result := F3(card1.Transactions, 2)
		b.StopTimer() // останавливаем таймер, чтобы время сравнения не учитывалось
		if !reflect.DeepEqual(result, want) {
			b.Fatalf("invalid result, got %v, want %v", result, want)
		}
		b.StartTimer() // продолжаем работу таймера
	}
}

// benchmark F4
func BenchmarkF4(b *testing.B) {
	card1 := &Card{ID: 1, Type: "Master", BankName: "Citi", CardNumber: "1111 2222 3333 4444", Balance: 20_000_00, CardDueDate: "2030-01-01"}
	card1.Transactions = MakeTransactions()

	want := map[string]int64{"Категория 3333": 100, "Категория 5555": 100000000, "Супермаркеты": 100000000}
	b.ResetTimer() // сбрасываем таймер, т.к. сама генерация транзакций достаточно ресурсоёмка
	for i := 0; i < b.N; i++ {
		result := F4(card1.Transactions, 2)
		b.StopTimer() // останавливаем таймер, чтобы время сравнения не учитывалось
		if !reflect.DeepEqual(result, want) {
			b.Fatalf("invalid result, got %v, want %v", result, want)
		}
		b.StartTimer() // продолжаем работу таймера
	}
}
*/
