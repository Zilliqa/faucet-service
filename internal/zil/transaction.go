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

func BatchConfirmer(
	provider *provider.Provider,
) func([]string) ([]bool, error) {
	return func(txIDs []string) ([]bool, error) {
		responses, err := provider.GetTransactionBatch(txIDs)
		if err != nil {
			return nil, err
		}
		result := []bool{}
		for _, v := range responses {
			result = append(result, v.Receipt.Success)
		}
		return result, nil
	}
}

func BatchSender(
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
		result, err := wallet.SendBatchOneGo(txs, *provider)
		if err != nil {
			return nil, err
		}

		txIDs := []string{}
		for _, v := range result {
			txIDs = append(txIDs, v.Hash)
		}
		return &txIDs, nil
	}
}
