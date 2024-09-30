package walletAndOkx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
)

// Структура для хранения информации о транзакции
type WalletTransaction struct {
	Amount      string
	ToAddress   string
	FromAddress string
	Status      string
}

// Структура для SDK опций (например, адрес, провайдер и т.д.)
type SDKOptions struct {
	Address  string
	Provider string // URL RPC провайдера (TON или другой)
	Account  string // Адрес аккаунта
}

// Структура для запроса отправки транзакции на блокчейн TON
type TransferRequest struct {
	Recipient string `json:"recipient"`
	Amount    string `json:"amount"`
}

// Функция для проверки баланса токенов на кошельке
func CheckWalletBalance(address string) (*big.Int, error) {
	// Пример запроса к RPC для проверки баланса
	rpcURL := "https://toncenter.com/api/v2/jsonRPC"
	requestBody := map[string]interface{}{
		"method":  "getAddressInformation",
		"params":  []interface{}{address},
		"id":      1,
		"jsonrpc": "2.0",
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка при формировании запроса: %v", err)
	}

	req, err := http.NewRequest("POST", rpcURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании запроса: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	// Распознаем ответ и извлекаем баланс
	var jsonResponse map[string]interface{}
	err = json.Unmarshal(respBody, &jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("ошибка при разборе ответа: %v", err)
	}

	result, ok := jsonResponse["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("неверный формат ответа")
	}

	balanceStr, ok := result["balance"].(string)
	if !ok {
		return nil, fmt.Errorf("баланс не найден в ответе")
	}

	// Преобразуем строку в big.Int
	balance := new(big.Int)
	balance.SetString(balanceStr, 10)

	return balance, nil
}

// Функция для отправки токенов с кошелька на биржу OKX
func TransferTokensToOKX(transaction WalletTransaction, sdkOptions SDKOptions, valueToWithdrawal *big.Int) error {
	// Проверка доступности сети перед транзакцией
	fmt.Println("Проверка доступности сети для вывода средств...")
	// Пример: вызов функции проверки сети (mock)
	networkAvailable := true // заменить на реальную проверку сети
	if !networkAvailable {
		return fmt.Errorf("сеть для вывода средств недоступна")
	}

	// Вычитаем комиссии за транзакцию (пример с использованием мок данных)
	fee := big.NewInt(1000000) // Устанавливаем фиксированную комиссию (заменить на реальную)
	finalAmount := new(big.Int).Sub(valueToWithdrawal, fee)

	// Формируем запрос для отправки транзакции
	txRequest := TransferRequest{
		Recipient: transaction.ToAddress,
		Amount:    finalAmount.String(),
	}

	body, err := json.Marshal(txRequest)
	if err != nil {
		return fmt.Errorf("ошибка при формировании тела транзакции: %v", err)
	}

	// Отправляем транзакцию через RPC провайдер TON
	rpcURL := sdkOptions.Provider
	req, err := http.NewRequest("POST", rpcURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("ошибка при создании запроса для отправки транзакции: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка при отправке транзакции: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	fmt.Println("Ответ от RPC провайдера TON:", string(respBody))

	// Проверка успешности транзакции
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("транзакция не удалась, статус-код: %d", resp.StatusCode)
	}

	// Логирование успешной транзакции
	transaction.Amount = finalAmount.String()
	fmt.Printf("Транзакция на %s токенов успешно выполнена на адрес %s\n", transaction.Amount, transaction.ToAddress)

	return nil
}

// Основная функция выполнения перевода
func ExecuteTransfer(transaction WalletTransaction, sdkOptions SDKOptions) error {
	fmt.Println("Начинаем процесс перевода токенов с кошелька на биржу OKX...")

	// Шаг 1: Проверка баланса на кошельке
	balance, err := CheckWalletBalance(sdkOptions.Address)
	if err != nil {
		return fmt.Errorf("не удалось проверить баланс: %v", err)
	}

	fmt.Printf("Текущий баланс на кошельке: %s\n", balance.String())

	// Шаг 2: Вычисление суммы для вывода
	reserveAmount := big.NewInt(100000000000000000) // Устанавливаем резерв для кошелька (заменить на нужное значение)
	valueToWithdrawal := new(big.Int).Sub(balance, reserveAmount)

	// Шаг 3: Отправка токенов на биржу OKX
	err = TransferTokensToOKX(transaction, sdkOptions, valueToWithdrawal)
	if err != nil {
		return fmt.Errorf("ошибка при отправке токенов на OKX: %v", err)
	}

	return nil
}
