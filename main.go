package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/cloudflare/cloudflare-go/v4/zones"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		// log request to std
		log.Printf("Request received: %v", c.Request)

		token := c.Query("token")
		email := c.Query("email")
		zoneName := c.Query("zone")
		ipv4 := c.Query("ipv4")
		ipv6 := c.Query("ipv6")

		// log request parameters
		log.Printf("Token: %s", token)
		log.Printf("Email: %s", email)
		log.Printf("Zone: %s", zoneName)
		log.Printf("IPv4: %s", ipv4)
		log.Printf("IPv6: %s", ipv6)

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

		client := cloudflare.NewClient(
			option.WithAPIToken(token),
			option.WithAPIEmail(email),
		)

		// list zones with name leoheller.de
		zoneListParams := zones.ZoneListParams{
			Name: cloudflare.F(zoneName),
		}

		zoneList, err := client.Zones.List(context.TODO(), zoneListParams)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("%+v\n", zoneList)

		if len(zoneList.Result) == 0 {
			panic("No zone found")
		}

		zoneID := zoneList.Result[0].ID
		fmt.Printf("Zone ID: %s\n", zoneID)

		// First, try to list existing AAAA records
		records, err := client.DNS.Records.List(context.TODO(), zoneID, dns.ListRecordParams{
			Type: cloudflare.F(string(dns.AAAARecordTypeAAAA)),
			Name: cloudflare.F(zoneName),
		})
		if err != nil {
			log.Printf("Error listing DNS records: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
			return
		}

		if len(records.Result) == 0 {
			// No existing record, create new one
			_, err := client.DNS.Records.Create(
				context.TODO(),
				zoneID,
				dns.CreateRecordParams{
					Type:    cloudflare.F(string(dns.AAAARecordTypeAAAA)),
					Name:    cloudflare.F(zoneName),
					Content: cloudflare.F(ipv6),
					TTL:     cloudflare.F(1), // Auto TTL
				},
			)
			if err != nil {
				log.Printf("Error creating DNS record: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
				return
			}
		} else {
			// Update existing record
			recordID := records.Result[0].ID
			_, err := client.DNS.Records.Update(
				context.TODO(),
				zoneID,
				dns.UpdateRecordParams{
					ID:      recordID,
					Type:    cloudflare.F(string(dns.AAAARecordTypeAAAA)),
					Name:    cloudflare.F(zoneName),
					Content: cloudflare.F(ipv6),
					TTL:     cloudflare.F(1), // Auto TTL
				},
			)
			if err != nil {
				log.Printf("Error updating DNS record: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "DNS record updated successfully"})
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
