package main

import (
	"faucet-service/internal/util"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		url := c.Request.URL.String()
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		requestID := c.GetHeader("x-request-id")

		entry := log.WithFields(log.Fields{
			"status_code": statusCode,
			"method":      method,
			"url":         url,
			"latency":     latency,
			"client_iP":   clientIP,
			"user_agent":  userAgent,
			"request_id":  requestID,
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%d", statusCode)
			if statusCode >= http.StatusInternalServerError {
				entry.Error(msg)
			} else if statusCode >= http.StatusBadRequest {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}

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
	secret := os.Getenv("RECAPTCHA_SECRET")
	privKey := os.Getenv("PRIVATE_KEY")

	err := util.ValidateEnvVars(
		envType,
		secret,
		privKey,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Use release mode for staging and prod
	if envType != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	log.WithFields(log.Fields{"EnvType": envType}).Info("Start Faucet Service")

	r := gin.New()
	r.Use(cors(), logger(), gin.Recovery())
	r.GET("/livez", func(c *gin.Context) { c.String(http.StatusOK, "") })

	r.Run("0.0.0.0:8080")
}
