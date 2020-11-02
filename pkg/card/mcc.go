package card

const categoryNotFound = "Категория не найдена"

// TranslateMCC - TranslateMCC
func TranslateMCC(code string) string {
	mcc := map[string]string{
		"5411": "Супермаркеты",
		"5533": "Автоуслуги",
		"5912": "Аптеки",
		"1111": "Категория 1111",
		"3333": "Категория 3333",
		"5555": "Категория 5555",
	}
	if value, ok := mcc[code]; ok {
		return value
	}
	return categoryNotFound
}
