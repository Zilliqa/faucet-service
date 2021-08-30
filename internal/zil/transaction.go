//  Copyright (C) 2021 Zilliqa
//
//  This file is part of faucet-service.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package zil

import (
	"faucet-service/internal/faucet"

	"github.com/Zilliqa/gozilliqa-sdk/account"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/Zilliqa/gozilliqa-sdk/transaction"
)

func GetIsTxConfirmedFn(
	provider *provider.Provider,
) func(string) bool {
	return func(txID string) bool {
		_, err := provider.GetTransaction(txID)
		return err == nil
	}
}

func SendBatchTxFn(
	provider *provider.Provider,
	wallet *account.Wallet,
	amountInZil string,
	version string,
) func([]*faucet.FundRequest) (*[]string, error) {
	amount := amountInZil + "000000000000"
	return func(reqs []*faucet.FundRequest) (*[]string, error) {
		gasPrice, err := provider.GetMinimumGasPrice()
		if err != nil {
			return nil, err
		}

		txs := []*transaction.Transaction{}
		for _, cur := range reqs {
			tx := &transaction.Transaction{
				Version:  version,
				ToAddr:   cur.Address,
				Amount:   amount,
				GasPrice: gasPrice,
				GasLimit: "50",
				Code:     "",
				Data:     "",
				Priority: false,
			}
			txs = append(txs, tx)
		}

		err = wallet.SignBatch(txs, *provider)
		if err != nil {
			return nil, err
		}
		batchSendingResult, err := wallet.SendBatchOneGo(txs, *provider)
		if err != nil {
			return nil, err
		}

		txIDs := []string{}
		for _, v := range batchSendingResult {
			txIDs = append(txIDs, v.Hash)
		}
		return &txIDs, nil
	}
}
