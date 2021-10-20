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
	"net/http"
	"time"

	"github.com/Zilliqa/gozilliqa-sdk/bech32"
	"github.com/Zilliqa/gozilliqa-sdk/util"
	"github.com/Zilliqa/gozilliqa-sdk/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Model binding and validation
// https://github.com/gin-gonic/gin#model-binding-and-validation

type BodyParams struct {
	Address string `form:"address" json:"address" binding:"required"`
	Token   string `form:"token" json:"token" binding:"required"`
}

func Controller(
	verify func(*log.Entry, string, string) error,
	insert func(*FundRequest) error,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		body := BodyParams{}
		remoteIP := c.ClientIP()
		requestID := c.GetHeader("x-request-id")
		userAgent := c.Request.UserAgent()

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		address := body.Address

		// Support Bech32
		if validator.IsBech32(address) {
			address, _ = bech32.FromBech32Addr(address)
		}

		// Validate ByStr20 address
		if !validator.IsAddress(address) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address"})
			return
		}
		address = util.ToCheckSumAddress(address)

		reqLogger := log.WithFields(
			log.Fields{
				"request_id": requestID,
				"remote_ip":  remoteIP,
				"user_agent": userAgent,
			},
		)

		reqLogger.Info("body: ", body)

		err := verify(reqLogger, body.Token, remoteIP)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		item := FundRequest{
			ID:        uuid.New().String(),
			CreatedAt: time.Now().Format(time.RFC3339),
			Address:   address,
			TxID:      "",
		}

		err = insert(&item)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		reqLogger.Info(http.StatusOK)
		c.JSON(http.StatusOK, nil)
	}
}
