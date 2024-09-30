package walletAndOkx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Структура для хранения информации об аутентификации OKX
type OKXAuth struct {
	APIKey    string
	APISecret string
	Password  string
	Proxy     string
}

// Структура транзакции
type OKXTransaction struct {
	Amount    string
	ToAddress string
	Network   string
	Status    string
}

// Структура для запроса на вывод средств
type WithdrawRequest struct {
	Currency  string `json:"ccy"`
	Amount    string `json:"amt"`
	ToAddress string `json:"toAddr"`
	Network   string `json:"network"`
}

var okxAuth = OKXAuth{
	APIKey:    "your-api-key",
	APISecret: "your-api-secret",
	Password:  "your-password",
	Proxy:     "your-proxy",
}

var rpcURL = "https://toncenter.com/api/v2/jsonRPC" // Пример URL для TON RPC

// Функция для выполнения запроса на вывод средств с OKX
func WithdrawFromOKX(transaction OKXTransaction) error {
	url := "https://www.okx.com/api/v5/asset/withdrawal"
	reqBody := WithdrawRequest{
		Currency:  "TON", // Указываем валюту TON
		Amount:    transaction.Amount,
		ToAddress: transaction.ToAddress,
		Network:   transaction.Network,
	}

	// Формируем тело запроса в формате JSON
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("ошибка при формировании запроса: %v", err)
	}

	// Создаем HTTP-запрос
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("ошибка при создании запроса: %v", err)
	}

	// Добавляем необходимые заголовки для аутентификации
	req.Header.Set("OK-ACCESS-KEY", okxAuth.APIKey)
	req.Header.Set("OK-ACCESS-SIGN", okxAuth.APISecret) // Тут нужно добавить генерацию подписи
	req.Header.Set("OK-ACCESS-PASSPHRASE", okxAuth.Password)
	req.Header.Set("Content-Type", "application/json")

	// Отправляем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	// Читаем ответ от OKX
	bodyResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка при чтении ответа: %v", err)
	}
	fmt.Println("Ответ от OKX:", string(bodyResp))

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("вывод средств не удался, статус-код: %d", resp.StatusCode)
	}

	return nil
}

// Функция для проверки баланса TON-кошелька через RPC
func CheckTONBalance(address string) (float64, error) {
	// Пример запроса для проверки баланса через TON RPC
	requestBody := map[string]interface{}{
		"method":  "getAddressInformation",
		"params":  []interface{}{address},
		"id":      1,
		"jsonrpc": "2.0",
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return 0, fmt.Errorf("ошибка при формировании TON RPC запроса: %v", err)
	}

	req, err := http.NewRequest("POST", rpcURL, bytes.NewBuffer(body))
	if err != nil {
		return 0, fmt.Errorf("ошибка при создании TON RPC запроса: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("ошибка при отправке TON RPC запроса: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("ошибка при чтении TON RPC ответа: %v", err)
	}

	fmt.Println("TON RPC Ответ:", string(respBody))

	// Парсим ответ и извлекаем баланс
	var jsonResponse map[string]interface{}
	err = json.Unmarshal(respBody, &jsonResponse)
	if err != nil {
		return 0, fmt.Errorf("ошибка при разборе TON RPC ответа: %v", err)
	}

	// Извлечение баланса
	result, ok := jsonResponse["result"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("неверный формат ответа")
	}

	balanceStr, ok := result["balance"].(string)
	if !ok {
		return 0, fmt.Errorf("баланс не найден в ответе")
	}

	// Преобразуем строку в число
	var balance float64
	fmt.Sscanf(balanceStr, "%f", &balance)

	return balance, nil
}

// Функция для выполнения транзакции
func ExecuteTransaction(transaction OKXTransaction) error {
	fmt.Println("Начало транзакции с OKX на TON кошелек...")

	// Шаг 1: Вывод средств с OKX
	err := WithdrawFromOKX(transaction)
	if err != nil {
		return fmt.Errorf("вывод не удался: %v", err)
	}
	fmt.Println("Вывод средств успешен, ожидаем подтверждения...")

	// Шаг 2: Ожидание обновления баланса на TON кошельке
	for {
		balance, err := CheckTONBalance(transaction.ToAddress)
		if err != nil {
			return fmt.Errorf("не удалось проверить баланс TON: %v", err)
		}

		fmt.Printf("Текущий баланс на TON кошельке: %f\n", balance)
		if balance > 0 {
			fmt.Println("Депозит подтвержден!")
			transaction.Status = "Done"
			break
		}
		time.Sleep(60 * time.Second) // Ждем 60 секунд перед повторной проверкой баланса
	}

	return nil
}
