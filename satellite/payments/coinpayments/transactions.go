package coinpayments

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	cmdCreateTransaction      = "create_transaction"
	cmdGetTransactionInfo     = "get_tx_info"
	cmdGetTransactionInfoList = "get_tx_info_multi"
)

type Currency string

const (
	CurrencyUSD  Currency = "USD"
	CurrencyLTCT Currency = "LTCT"
)

func (c Currency) String() string {
	return string(c)
}

type TransactionID string

func (id TransactionID) String() string {
	return string(id)
}

type TransactionIDList []TransactionID

func (list TransactionIDList) String() string {
	if len(list) == 0 {
		return ""
	}
	if len(list) == 1 {
		return string(list[0])
	}

	var separator = "|"

	var builder strings.Builder
	builder.WriteString(string(list[0]))
	builder.WriteString(separator)

	for _, id := range list[1 : len(list)-1] {
		builder.WriteString(string(id))
		builder.WriteString(separator)
	}

	builder.WriteString(string(list[len(list)-1]))
	return builder.String()
}

type Transaction struct {
	ID             TransactionID
	Address        string
	Amount         float64
	DestTag        string
	ConfirmsNeeded int64
	Timeout        time.Duration
	CheckoutURL    string
	StatusURL      string
	QRCodeURL      string
}

func (tx *Transaction) UnmarshalJSON(b []byte) error {
	var txRaw struct {
		Amount         string `json:"amount"`
		Address        string `json:"address"`
		DestTag        string `json:"dest_tag"`
		TxID           string `json:"txn_id"`
		ConfirmsNeeded string `json:"confirms_needed"`
		Timeout        int    `json:"timeout"`
		CheckoutURL    string `json:"checkout_url"`
		StatusURL      string `json:"status_url"`
		QRCodeURL      string `json:"qrcode_url"`
	}

	if err := json.Unmarshal(b, &txRaw); err != nil {
		return err
	}

	amount, err := strconv.ParseFloat(txRaw.Amount, 64)
	if err != nil {
		return err
	}
	confirms, err := strconv.ParseInt(txRaw.ConfirmsNeeded, 10, 64)
	if err != nil {
		return err
	}

	*tx = Transaction{
		ID:             TransactionID(txRaw.TxID),
		Address:        txRaw.Address,
		Amount:         amount,
		DestTag:        txRaw.DestTag,
		ConfirmsNeeded: confirms,
		Timeout:        time.Second * time.Duration(txRaw.Timeout),
		CheckoutURL:    txRaw.CheckoutURL,
		StatusURL:      txRaw.StatusURL,
		QRCodeURL:      txRaw.QRCodeURL,
	}

	return nil
}

type TransactionInfo struct {
	Address          string
	Coin             Currency
	Amount           float64
	Received         float64
	ConfirmsReceived int
	Status           int
	ExpiresAt        time.Time
	CreatedAt        time.Time
}

func (info *TransactionInfo) UnmarshalJSON(b []byte) error {
	var txInfoRaw struct {
		Address      string `json:"payment_address"`
		Coin         string `json:"coin"`
		Status       int    `json:"status"`
		AmountF      string `json:"amountf"`
		ReceivedF    string `json:"receivedf"`
		ConfirmsRecv int    `json:"recv_confirms"`
		ExpiresAt    int64  `json:"time_expires"`
		CreatedAt    int64  `json:"time_created"`
	}

	if err := json.Unmarshal(b, &txInfoRaw); err != nil {
		return err
	}

	amount, err := strconv.ParseFloat(txInfoRaw.AmountF, 64)
	if err != nil {
		return err
	}
	received, err := strconv.ParseFloat(txInfoRaw.ReceivedF, 64)
	if err != nil {
		return err
	}

	*info = TransactionInfo{
		Address:          txInfoRaw.Address,
		Coin:             Currency(txInfoRaw.Coin),
		Amount:           amount,
		Received:         received,
		ConfirmsReceived: txInfoRaw.ConfirmsRecv,
		Status:           txInfoRaw.Status,
		ExpiresAt:        time.Unix(txInfoRaw.ExpiresAt, 0),
		CreatedAt:        time.Unix(txInfoRaw.CreatedAt, 0),
	}

	return nil
}

type TransactionInfos map[TransactionID]TransactionInfo

func (infos *TransactionInfos) UnmarshalJSON(b []byte) error {
	var _infos map[TransactionID]struct {
		Error string `json:"error"`
		TransactionInfo
	}

	if err := json.Unmarshal(b, &_infos); err != nil {
		return err
	}

	for id, info := range _infos {
		if info.Error != "ok" {
			return errors.New(info.Error)
		}

		(*infos)[id] = info.TransactionInfo
	}

	return nil
}

type CreateTX struct {
	Amount      float64
	CurrencyIn  Currency
	CurrencyOut Currency
	BuyerEmail  string
}

type Transactions struct {
	client *Client
}

func (c Transactions) Create(ctx context.Context, params CreateTX) (*Transaction, error) {
	amount := strconv.FormatFloat(params.Amount, 'f', -1, 64)

	values := make(url.Values)
	values.Set("amount", amount)
	values.Set("currency1", params.CurrencyIn.String())
	values.Set("currency2", params.CurrencyOut.String())
	values.Set("buyer_email", params.BuyerEmail)

	tx := new(Transaction)

	res, err := c.client.do(ctx, cmdCreateTransaction, values)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(res, tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (c Transactions) Info(ctx context.Context, id TransactionID) (*TransactionInfo, error) {
	values := make(url.Values)
	values.Set("txid", id.String())

	txInfo := new(TransactionInfo)

	res, err := c.client.do(ctx, cmdGetTransactionInfo, values)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(res, txInfo); err != nil {
		return nil, err
	}

	return txInfo, nil
}

func (c Transactions) ListInfos(ctx context.Context, ids TransactionIDList) (TransactionInfos, error) {
	if len(ids) > 25 {
		return nil, errors.New("only up to 25 transactions can be queried")
	}

	values := make(url.Values)
	values.Set("txid", ids.String())

	txInfos := make(TransactionInfos, len(ids))

	res, err := c.client.do(ctx, cmdGetTransactionInfoList, values)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(res, &txInfos); err != nil {
		return nil, err
	}

	return txInfos, nil
}
