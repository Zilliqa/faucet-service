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

package main

import (
	"faucet-service/internal/faucet"
	"faucet-service/internal/recaptcha"
	"faucet-service/internal/zil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Zilliqa/gozilliqa-sdk/account"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/Zilliqa/gozilliqa-sdk/util"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, Cache-Control, pragma, Expires, Origin, x-request-id")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Writer.Header().Set("pragma", "no-cache")
		c.Writer.Header().Set("expires", "0")
		c.Writer.Header().Set("x-content-type-options", "nosniff")
		c.Writer.Header().Set("x-frame-options", "DENY")
		c.Writer.Header().Set("x-xss-protection", "1; mode=block")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}

func main() {
	envType := os.Getenv("ENV_TYPE")
	nodeURL := os.Getenv("NODE_URL")
	chainIDStr := os.Getenv("CHAIN_ID")
	amountInZil := os.Getenv("AMOUNT_IN_ZIL")
	batchInterval := os.Getenv("BATCH_INTERVAL")
	batchLimitStr := os.Getenv("BATCH_LIMIT")
	ttlStr := os.Getenv("TTL")
	privKey := os.Getenv("PRIVATE_KEY")
	secret := os.Getenv("RECAPTCHA_SECRET")

	envVars := []string{
		envType,
		nodeURL,
		chainIDStr,
		amountInZil,
		batchInterval,
		batchLimitStr,
		ttlStr,
		privKey,
		secret,
	}
	for _, envVar := range envVars {
		if envVar == "" {
			panic("💥Invalid environment variables")
		}
	}
	chainID, err := strconv.Atoi(chainIDStr)
	if err != nil {
		panic(err)
	}
	ttl, err := strconv.Atoi(ttlStr)
	if err != nil {
		panic(err)
	}
	batchLimit, err := strconv.Atoi(batchLimitStr)
	if err != nil {
		panic(err)
	}

	// Create the DB schema
	mdb, err := faucet.NewMemDB("req")
	if err != nil {
		panic(err)
	}

	// Use release mode for staging and prod
	if envType != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	logger := log.WithFields(log.Fields{"EnvType": envType})

	wallet := account.NewWallet()
	wallet.AddByPrivateKey(privKey)
	curProvider := provider.NewProvider(nodeURL)
	msgVersion := 1
	version := strconv.FormatInt(int64(util.Pack(chainID, msgVersion)), 10)

	logger.Infof("🚀NodeURL:%v ChainID:%v Amount:%v BatchInterval:%v",
		nodeURL,
		chainID,
		amountInZil,
		batchInterval,
	)

	batchConfirmTx := zil.NewBatchConfirmTxFn(curProvider)
	batchSendTx := zil.NewBatchSendTxFn(
		curProvider,
		wallet,
		amountInZil,
		version,
	)

	// Funcs are invoked in their own goroutine, asynchronously.
	c := cron.New()
	c.AddFunc("@every 10s", func() {
		total, totalReq, totalTx, err := mdb.Scan()
		if err != nil {
			logger.Error(err)
			return
		}
		logger.Infof("📡Total:%d Req:%d Tx:%d",
			total,
			totalReq,
			totalTx,
		)
	})

	c.AddFunc("@every "+batchInterval, func() {

		t0 := time.Now()
		// Deletes the confirmed items which are no longer needed.
		countConfirmed, err := mdb.Confirm(batchConfirmTx, batchLimit)
		if err != nil {
			logger.Error(err)
			return
		}
		logger.Infof("✅Confirmed:%d (%v)", countConfirmed, time.Since(t0))

		t0 = time.Now()
		// Reduce stored data volumes by expiring the old items.
		// which are either pending or unconfirmed.
		now := time.Now().Unix()
		countExpired, err := mdb.Expire(now, ttl)
		if err != nil {
			logger.Error(err)
			return
		}
		logger.Infof("⌛️Expired:%d (%v)", countExpired, time.Since(t0))

		t0 = time.Now()
		// Retry unconfirmed items by removing the old tx id.
		// Note that it's at-least-once delivery.
		countRetry, err := mdb.Retry()
		if err != nil {
			logger.Error(err)
			return
		}
		logger.Infof("🔸Retry:%d (%v)", countRetry, time.Since(t0))

		t0 = time.Now()
		// Send transactions
		countBatch, err := mdb.Send(batchSendTx, batchLimit)
		if err != nil {
			logger.Error(err)
			return
		}
		logger.Infof("🔹Batch:%d (%v)", countBatch, time.Since(t0))

	})
	c.Start()

	r := gin.New()
	r.Use(cors(), gin.Recovery())

	r.GET("/livez", func(c *gin.Context) { c.String(http.StatusOK, "") })

	verifyToken := recaptcha.VerifyToken
	// Mock verifyToken for dev
	if envType == "dev" {
		verifyToken = func(l *log.Entry, url string) error {
			return nil
		}
	}

	r.POST("api/v1/faucet", faucet.Controller(
		secret,
		verifyToken,
		mdb.Insert,
	))

	r.Run("0.0.0.0:8080")
}
