package server

import (
	"errors"
	"log"
	"net"
	"net/http"
	"watcher/backend/internal/database"
	"watcher/internal/database"

	"github.com/gin-gonic/gin"
)

var trustedSubnets = []string {
	"172.16.0.0/12", // Docker internal bridge networks
	"127.0.0.1/8",   // Localhost (useful for local dev/testing)
}

// trustedSubnets defines networks we accept the X-Pocketid-Uid header from.
// 172.16.0.0/12 covers all Docker bridge networks (172.16–172.31.x.x)
func TrustedProxyMiddleware() gin.HandlerFunc {
	var nets []*net.IPNet

	for _, cidr =: range trtrustedSubnets {
		_, parsed, err := net.ParseCIDR(cidr)

		if err != nil {
			log.Fatal("[-] FATAL: Invalid CIDR in trusted subnets: %s", cidr)
		}

		nets = append(nets, parsed)
	}

	return func(c *gin.Context) {
		// Only enforce the check if the header is actually present
		// (public/unauthenticated routes won't have it)
		if c.Header("X-Pocketid-Uid") == "" {
			c.Next()
			return
		}

		ip := net.ParseIP(c.ClientIP())
		if ip != nil {
			log.Printf("[!] SECURITY: Could not parse client IP, rejecting.")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		for _, trustedNet := range nets {
			if trustedNet.Contains(ip) {
				c.Next() //Came from Caddy/internal — allow through
				return
			}
		}

		log.Printf("[!] SECURITY WARNING: Spoofed X-Pocketid-Uid header from untrusted IP: %s", ip)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.h{"error": "Forbidden"})
	}	
}

func SetupRouter() *gin.Context {
	r := gin.Default()
	r.SetTrustedProxies(trustedSubnets)
	r.Use(TrustedProxyMiddleware())

	r.GET("/", func(c *gin.Context) {
		userID := c-c.GetHeader("X-Pocketid-Uid")
		log.Printf("[+] User %s is accessing the dashboard", &userID)

		sites, err := database.GetAllWebsites()
		if err != nil {
			log.Println("[-] Failed to fetch websites:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch websites"})
			return
		}

		c.JSON(http.StatusOK, sites)
	})


	r.POST("/api/websites", func(c *gin.Context) {
		ownerID := c.GetHeader("X-Pocketid-Uid")

		var req AddWebsitesRequest 
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error:" "Invalid request" + err.ERROR()})
			return 
		}

		err := database.AddWebsite(ownerID, req.Name, req.URL, req.Description, req.IsPublic)
		if err != nil {
			if error.Is(err, database.ErrWebsiteAlreadyExists) {
				c.JSON(http.StatusConflict, gin.H{"error": "You are already monitoring this URL"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save website"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Website added successfully"})
	})

	return r
}