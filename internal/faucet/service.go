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

package faucet

import (
	"time"

	"github.com/hashicorp/go-memdb"
)

type FundRequest struct {
	ID        string
	CreatedAt string
	Address   string
	TxID      string
}

type MDB struct {
	DB        *memdb.MemDB
	TableName string
}

func NewMemDB(tableName string) (*MDB, error) {
	mdb := MDB{}
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			(tableName): {
				Name: tableName,
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"created_at": {
						Name:    "created_at",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "CreatedAt"},
					},
				},
			},
		},
	}
	newDb, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, err
	}
	mdb.DB = newDb
	mdb.TableName = tableName
	return &mdb, nil
}

func (mdb *MDB) Insert(item *FundRequest) error {
	// Create a write transaction
	db := mdb.DB
	tableName := mdb.TableName
	txn := db.Txn(true)

	if err := txn.Insert(tableName, item); err != nil {
		txn.Abort()
		return err
	}

	// Commit the transaction
	txn.Commit()
	return nil
}

func (mdb *MDB) Scan() (int, int, int, error) {
	db := mdb.DB
	tableName := mdb.TableName

	txn := db.Txn(false)
	defer txn.Abort()

	total := 0
	totalReq := 0
	totalTx := 0

	it, err := txn.Get(tableName, "id")
	if err != nil {
		return total, totalReq, totalTx, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		cur := obj.(*FundRequest)

		total++
		if cur.TxID == "" {
			totalReq++
		} else {
			totalTx++
		}
	}

	return total, totalReq, totalTx, nil
}

func (mdb *MDB) Confirm(
	batchConfirmTx func([]string) ([]bool, error),
	batchLimit int,
) (int, error) {
	db := mdb.DB
	tableName := mdb.TableName

	count := 0
	txn := db.Txn(true)
	it, err := txn.Get(tableName, "id")
	if err != nil {
		txn.Abort()
		return count, err
	}

	reqs := []*FundRequest{}
	txIDs := []string{}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		if len(txIDs) >= batchLimit {
			break
		}
		cur := obj.(*FundRequest)
		if cur.TxID != "" {
			reqs = append(reqs, cur)
			txIDs = append(txIDs, cur.TxID)
		}
	}
	if len(txIDs) == 0 {
		txn.Abort()
		return count, nil
	}

	result, err := batchConfirmTx(txIDs)
	if err != nil {
		txn.Abort()
		return count, err
	}

	for i, v := range result {
		if v {
			txn.Delete(tableName, reqs[i])
			count++
		}
	}

	txn.Commit()
	return count, nil
}

func (mdb *MDB) Expire(now int64, ttl int) (int, error) {
	db := mdb.DB
	tableName := mdb.TableName

	count := 0
	txn := db.Txn(true)
	it, err := txn.Get(tableName, "id")
	if err != nil {
		txn.Abort()
		return count, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		cur := obj.(*FundRequest)

		t, _ := time.Parse(time.RFC3339, cur.CreatedAt)
		createdAt := t.Unix()

		isExpired := int(createdAt) < int(now)-ttl
		if isExpired {
			txn.Delete(tableName, cur)
			count++
		}
	}
	txn.Commit()
	return count, nil
}

func (mdb *MDB) Retry() (int, error) {
	db := mdb.DB
	tableName := mdb.TableName

	count := 0
	txn := db.Txn(true)
	it, err := txn.Get(tableName, "id")
	if err != nil {
		txn.Abort()
		return count, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		cur := obj.(*FundRequest)

		// Note that items are the fresh ones here as we assume that
		// Expire() is executed before Retry()
		//
		// Retry the batch job for the FRESH and UNCONFIRMED items.
		// The ones with the tx id are unconfirmed one as we assume that
		// Confirm() is executed before Retry(),
		hasTxID := cur.TxID != ""
		if hasTxID {
			newItem := &FundRequest{
				ID:        cur.ID,
				CreatedAt: cur.CreatedAt,
				Address:   cur.Address,
				TxID:      "",
			}
			txn.Insert(tableName, newItem)
			count++
		}
	}
	txn.Commit()
	return count, nil
}

func (mdb *MDB) Send(
	sendBatchTx func([]*FundRequest) (*[]string, error),
	batchLimit int,
) (int, error) {
	db := mdb.DB
	tableName := mdb.TableName

	count := 0
	txn := db.Txn(true)
	it, err := txn.Get(tableName, "id")
	if err != nil {
		txn.Abort()
		return count, err
	}

	reqsWithoutTxID := []*FundRequest{}
	// Filter items without txID

	for obj := it.Next(); obj != nil; obj = it.Next() {
		if len(reqsWithoutTxID) >= batchLimit {
			break
		}

		cur := obj.(*FundRequest)
		if cur.TxID == "" {
			reqsWithoutTxID = append(reqsWithoutTxID, cur)

		}
	}

	if len(reqsWithoutTxID) == 0 {
		txn.Abort()
	} else {
		txIDs, err := sendBatchTx(reqsWithoutTxID)
		if err != nil {
			txn.Abort()
			return count, err
		}
		for i, cur := range reqsWithoutTxID {
			count++
			// When updating an object, the obj provided should be a copy
			// rather than a value updated in-place.
			newItem := &FundRequest{
				ID:        cur.ID,
				CreatedAt: cur.CreatedAt,
				Address:   cur.Address,
				TxID:      (*txIDs)[i],
			}
			txn.Insert(tableName, newItem)
		}
		txn.Commit()
	}
	return count, nil
}
