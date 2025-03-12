package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/dns"
	"github.com/cloudflare/cloudflare-go/v3/option"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		token := c.Query("token")
		email := c.Query("email")
		zoneName := c.Query("zone")
		recordName := c.Query("record")
		ipv4 := c.Query("ipv4")
		ipv6 := c.Query("ipv6")

		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Missing token URL parameter."})
			return
		}
		if email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Missing email URL parameter."})
			return
		}
		if zoneName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Missing zone URL parameter."})
			return
		}
		if ipv4 == "" && ipv6 == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Missing ipv4 or ipv6 URL parameter."})
			return
		}

		api, err := cloudflare.NewClient(
			option.WithAPIKey(token),
			option.WithAPIEmail(email),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
			return
		}

		ctx := context.Background()

		zoneID, err := api.Zones.GetZoneIDByName(ctx, zoneName)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Zone does not exist."})
			return
		}

		recordFullName := zoneName
		if recordName != "" {
			recordFullName = recordName + "." + zoneName
		}

		if ipv4 != "" {
			err := updateDNSRecord(ctx, api, zoneID, recordFullName, dns.RecordTypeA, ipv4)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
				return
			}
		}

		if ipv6 != "" {
			err := updateDNSRecord(ctx, api, zoneID, recordFullName, dns.RecordTypeAAAA, ipv6)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Update successful."})
	})

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "OK"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func updateDNSRecord(ctx context.Context, api *cloudflare.API, zoneID, recordName, recordType, content string) error {
	records, err := api.DNS.Records.List(ctx, zoneID, dns.RecordListParams{
		Name: &recordName,
		Type: &recordType,
	})
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return fmt.Errorf("%s record for %s does not exist", recordType, recordName)
	}

	record := records[0]
	if record.Content != content {
		_, err := api.DNS.Records.Update(ctx, zoneID, dns.RecordUpdateParams{
			ID: &record.ID,
			Record: dns.Record{
				Type:    &recordType,
				Name:    &recordName,
				Content: &content,
				TTL:     record.TTL,
				Proxied: record.Proxied,
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
