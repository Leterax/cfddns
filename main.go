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
		recordName := c.Query("record")
		ipv4 := c.Query("ipv4")
		ipv6 := c.Query("ipv6")

		// log request parameters
		log.Printf("Token: %s", token)
		log.Printf("Email: %s", email)
		log.Printf("Zone: %s", zoneName)
		log.Printf("Record: %s", recordName)
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
			option.WithAPIKey(token), // defaults to os.LookupEnv("CLOUDFLARE_API_KEY")
			option.WithAPIEmail(email),               // defaults to os.LookupEnv("CLOUDFLARE_EMAIL")
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
		
		recordResponse, err := client.DNS.Records.Update(
			context.TODO(),
			zoneID,
			dns.RecordUpdateParams{
				ZoneID: cloudflare.F(zoneID),
				Record: dns.ARecordParam{
					Content: cloudflare.F(ipv6),
					Name:    cloudflare.F(recordName),
					Type:    cloudflare.F(dns.ARecordType(dns.AAAARecordTypeAAAA)),
				},
			},
		)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("%+v\n", recordResponse)

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
