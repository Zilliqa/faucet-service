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
	"errors"
	"testing"
	"time"
)

func TestScan(t *testing.T) {
	mdb, err := NewMemDB("test")
	if err != nil {
		panic(err)
	}

	testCases := []struct {
		items     []*FundRequest
		wantTotal int
		wantReq   int
		wantTx    int
		wantErr   error
	}{
		{
			[]*FundRequest{},
			0,
			0,
			0,
			nil,
		},
	}
	for _, testCase := range testCases {
		total, req, tx, err := mdb.Scan()
		if total != testCase.wantTotal {
			t.Errorf("%#v; want %#v", total, testCase.wantTotal)
		}
		if req != testCase.wantReq {
			t.Errorf("%#v; want %#v", req, testCase.wantReq)
		}
		if tx != testCase.wantTx {
			t.Errorf("%#v; want %#v", tx, testCase.wantTx)
		}
		if err != testCase.wantErr {
			t.Errorf("%#v; want %#v", err, testCase.wantErr)
		}
	}

	samples := []*FundRequest{
		{
			ID:        "1e1413cc-f604-44db-969d-eb3b40fea4a1",
			CreatedAt: "2021-08-29T04:00:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "b56281dc5dd8b44f37fc44bb12d3cc170616eeef121abd364369b15b9b8473a3",
		},
		{
			ID:        "22db01c2-2e3b-4ec2-9ce7-77f0372a50b2",
			CreatedAt: "2021-08-29T04:00:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "9bb742ccb83de0689fee2e8f7967ced7e0d60f1577d94a08c770e8e49838e187",
		},
		{
			ID:        "1e1413cc-f604-44db-969d-eb3b40fea4a3",
			CreatedAt: "2021-08-29T04:00:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "",
		},
	}
	for _, v := range samples {
		mdb.Insert(v)
	}

	testCases = []struct {
		items     []*FundRequest
		wantTotal int
		wantReq   int
		wantTx    int
		wantErr   error
	}{
		{
			[]*FundRequest{},
			3,
			1,
			2,
			nil,
		},
	}
	for _, testCase := range testCases {
		total, req, tx, err := mdb.Scan()
		if total != testCase.wantTotal {
			t.Errorf("%#v; want %#v", total, testCase.wantTotal)
		}
		if req != testCase.wantReq {
			t.Errorf("%#v; want %#v", req, testCase.wantReq)
		}
		if tx != testCase.wantTx {
			t.Errorf("%#v; want %#v", tx, testCase.wantTx)
		}
		if err != testCase.wantErr {
			t.Errorf("%#v; want %#v", err, testCase.wantErr)
		}
	}
}

func TestConfirm(t *testing.T) {
	mdb, err := NewMemDB("test")
	if err != nil {
		panic(err)
	}
	// Mock batchConfirmTx()
	batchConfirmTx := func(txIDs []string) ([]bool, error) {
		result := []bool{}
		for range txIDs {
			result = append(result, true)
		}
		return result, nil
	}

	count, _ := mdb.Confirm(batchConfirmTx, 10)
	if count != 0 {
		t.Errorf("%#v", count)
	}

	samples := []*FundRequest{
		{
			ID:        "1e1413cc-f604-44db-969d-eb3b40fea4a0",
			CreatedAt: "2021-08-29T04:00:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "b56281dc5dd8b44f37fc44bb12d3cc170616eeef121abd364369b15b9b8473a1",
		},
		{
			ID:        "1e1413cc-f604-44db-969d-eb3b40fea4a1",
			CreatedAt: "2021-08-29T04:00:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "b56281dc5dd8b44f37fc44bb12d3cc170616eeef121abd364369b15b9b8473a2",
		},
		{
			ID:        "22db01c2-2e3b-4ec2-9ce7-77f0372a50b2",
			CreatedAt: "2021-08-29T04:00:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "9bb742ccb83de0689fee2e8f7967ced7e0d60f1577d94a08c770e8e49838e183",
		},
		{
			ID:        "1e1413cc-f604-44db-969d-eb3b40fea4a3",
			CreatedAt: "2021-08-29T04:00:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "",
		},
	}
	for _, v := range samples {
		mdb.Insert(v)
	}

	mdb.Confirm(batchConfirmTx, 2)

	testCases := []struct {
		items     []*FundRequest
		wantTotal int
		wantReq   int
		wantTx    int
		wantErr   error
	}{
		{
			[]*FundRequest{},
			2,
			1,
			1,
			nil,
		},
	}
	for _, testCase := range testCases {
		total, req, tx, err := mdb.Scan()
		if total != testCase.wantTotal {
			t.Errorf("%#v; want %#v", total, testCase.wantTotal)
		}
		if req != testCase.wantReq {
			t.Errorf("%#v; want %#v", req, testCase.wantReq)
		}
		if tx != testCase.wantTx {
			t.Errorf("%#v; want %#v", tx, testCase.wantTx)
		}
		if err != testCase.wantErr {
			t.Errorf("%#v; want %#v", err, testCase.wantErr)
		}
	}
}

func TestCleanup(t *testing.T) {
	mdb, err := NewMemDB("test")
	if err != nil {
		panic(err)
	}
	samples := []*FundRequest{
		{
			ID:        "1e1413cc-f604-44db-969d-eb3b40fea4a1",
			CreatedAt: "2021-08-29T04:00:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "",
		},
		{
			ID:        "22db01c2-2e3b-4ec2-9ce7-77f0372a50b2",
			CreatedAt: "2021-08-29T04:02:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "",
		},
		{
			ID:        "1e1413cc-f604-44db-969d-eb3b40fea4a3",
			CreatedAt: "2021-08-29T04:04:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "",
		},
	}

	for _, v := range samples {
		mdb.Insert(v)
	}

	now, _ := time.Parse(time.RFC3339, "2021-08-29T04:04:07Z")
	nowUnix := now.Unix()
	ttl := 60 * 3
	mdb.Expire(nowUnix, ttl)

	testCases := []struct {
		items     []*FundRequest
		wantTotal int
		wantReq   int
		wantTx    int
		wantErr   error
	}{
		{
			[]*FundRequest{},
			2,
			2,
			0,
			nil,
		},
	}
	for _, testCase := range testCases {
		total, req, tx, err := mdb.Scan()
		if total != testCase.wantTotal {
			t.Errorf("%#v; want %#v", total, testCase.wantTotal)
		}
		if req != testCase.wantReq {
			t.Errorf("%#v; want %#v", req, testCase.wantReq)
		}
		if tx != testCase.wantTx {
			t.Errorf("%#v; want %#v", tx, testCase.wantTx)
		}
		if err != testCase.wantErr {
			t.Errorf("%#v; want %#v", err, testCase.wantErr)
		}
	}
}

func TestRetry(t *testing.T) {
	mdb, err := NewMemDB("test")
	if err != nil {
		panic(err)
	}
	samples := []*FundRequest{
		{
			ID:        "1e1413cc-f604-44db-969d-eb3b40fea4a1",
			CreatedAt: "2021-08-29T04:00:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "b56281dc5dd8b44f37fc44bb12d3cc170616eeef121abd364369b15b9b8473a1",
		},
		{
			ID:        "22db01c2-2e3b-4ec2-9ce7-77f0372a50b2",
			CreatedAt: "2021-08-29T04:02:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "b56281dc5dd8b44f37fc44bb12d3cc170616eeef121abd364369b15b9b8473a2",
		},
		{
			ID:        "1e1413cc-f604-44db-969d-eb3b40fea4a3",
			CreatedAt: "2021-08-29T04:04:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "",
		},
	}

	for _, v := range samples {
		mdb.Insert(v)
	}

	mdb.Retry()

	testCases := []struct {
		items     []*FundRequest
		wantTotal int
		wantReq   int
		wantTx    int
		wantErr   error
	}{
		{
			[]*FundRequest{},
			3,
			3,
			0,
			nil,
		},
	}
	for _, testCase := range testCases {
		total, req, tx, err := mdb.Scan()
		if total != testCase.wantTotal {
			t.Errorf("%#v; want %#v", total, testCase.wantTotal)
		}
		if req != testCase.wantReq {
			t.Errorf("%#v; want %#v", req, testCase.wantReq)
		}
		if tx != testCase.wantTx {
			t.Errorf("%#v; want %#v", tx, testCase.wantTx)
		}
		if err != testCase.wantErr {
			t.Errorf("%#v; want %#v", err, testCase.wantErr)
		}
	}
}

func TestBatchErr(t *testing.T) {
	mdb, err := NewMemDB("test")
	if err != nil {
		panic(err)
	}
	samples := []*FundRequest{
		{
			ID:        "1e1413cc-f604-44db-969d-eb3b40fea4a1",
			CreatedAt: "2021-08-29T04:00:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "",
		},
		{
			ID:        "22db01c2-2e3b-4ec2-9ce7-77f0372a50b2",
			CreatedAt: "2021-08-29T04:02:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "",
		},
		{
			ID:        "1e1413cc-f604-44db-969d-eb3b40fea4a3",
			CreatedAt: "2021-08-29T04:04:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "",
		},
	}

	for _, v := range samples {
		mdb.Insert(v)
	}

	// Mock sendBatchTx()
	sendBatchTxErr := func(reqs []*FundRequest) (*[]string, error) {
		return nil, errors.New("Negative Testing")
	}

	_, err = mdb.Send(sendBatchTxErr, 10)
	if err.Error() != "Negative Testing" {
		t.Errorf("%#v", err.Error())
	}
}
func TestBatchNoop(t *testing.T) {
	mdb, err := NewMemDB("test")
	if err != nil {
		panic(err)
	}
	samples := []*FundRequest{
		{
			ID:        "1e1413cc-f604-44db-969d-eb3b40fea4a1",
			CreatedAt: "2021-08-29T04:00:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "b56281dc5dd8b44f37fc44bb12d3cc170616eeef121abd364369b15b9b8473a1",
		},
		{
			ID:        "22db01c2-2e3b-4ec2-9ce7-77f0372a50b2",
			CreatedAt: "2021-08-29T04:02:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "b56281dc5dd8b44f37fc44bb12d3cc170616eeef121abd364369b15b9b8473a2",
		},
		{
			ID:        "1e1413cc-f604-44db-969d-eb3b40fea4a3",
			CreatedAt: "2021-08-29T04:04:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "b56281dc5dd8b44f37fc44bb12d3cc170616eeef121abd364369b15b9b8473a3",
		},
	}

	for _, v := range samples {
		mdb.Insert(v)
	}

	// Mock sendBatchTx()
	sendBatchTx := func(reqs []*FundRequest) (*[]string, error) {
		txIDs := []string{}
		for _, v := range reqs {
			txIDs = append(txIDs, v.ID)
		}
		return &txIDs, nil
	}

	mdb.Send(sendBatchTx, 10)

	testCases := []struct {
		items     []*FundRequest
		wantTotal int
		wantReq   int
		wantTx    int
		wantErr   error
	}{
		{
			[]*FundRequest{},
			3,
			0,
			3,
			nil,
		},
	}
	for _, testCase := range testCases {
		total, req, tx, err := mdb.Scan()
		if total != testCase.wantTotal {
			t.Errorf("%#v; want %#v", total, testCase.wantTotal)
		}
		if req != testCase.wantReq {
			t.Errorf("%#v; want %#v", req, testCase.wantReq)
		}
		if tx != testCase.wantTx {
			t.Errorf("%#v; want %#v", tx, testCase.wantTx)
		}
		if err != testCase.wantErr {
			t.Errorf("%#v; want %#v", err, testCase.wantErr)
		}
	}
}

func TestBatch(t *testing.T) {
	mdb, err := NewMemDB("test")
	if err != nil {
		panic(err)
	}
	samples := []*FundRequest{
		{
			ID:        "1e1413cc-f604-44db-969d-eb3b40fea4a1",
			CreatedAt: "2021-08-29T04:00:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "",
		},
		{
			ID:        "22db01c2-2e3b-4ec2-9ce7-77f0372a50b2",
			CreatedAt: "2021-08-29T04:02:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "",
		},
		{
			ID:        "1e1413cc-f604-44db-969d-eb3b40fea4a3",
			CreatedAt: "2021-08-29T04:04:07Z",
			Address:   "0x0334995e2CFc53CF785C554839F6e845A3A09e79",
			TxID:      "",
		},
	}

	for _, v := range samples {
		mdb.Insert(v)
	}

	// Mock sendBatchTx()
	sendBatchTx := func(reqs []*FundRequest) (*[]string, error) {
		txIDs := []string{}
		for _, v := range reqs {
			txIDs = append(txIDs, v.ID)
		}
		return &txIDs, nil
	}
	mdb.Send(sendBatchTx, 2)

	testCases := []struct {
		items     []*FundRequest
		wantTotal int
		wantReq   int
		wantTx    int
		wantErr   error
	}{
		{
			[]*FundRequest{},
			3,
			1,
			2,
			nil,
		},
	}
	for _, testCase := range testCases {
		total, req, tx, err := mdb.Scan()
		if total != testCase.wantTotal {
			t.Errorf("%#v; want %#v", total, testCase.wantTotal)
		}
		if req != testCase.wantReq {
			t.Errorf("%#v; want %#v", req, testCase.wantReq)
		}
		if tx != testCase.wantTx {
			t.Errorf("%#v; want %#v", tx, testCase.wantTx)
		}
		if err != testCase.wantErr {
			t.Errorf("%#v; want %#v", err, testCase.wantErr)
		}
	}
}
