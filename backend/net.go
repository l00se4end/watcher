package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// Monitor struct za bazu i template
type Monitor struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Status string `json:"status"`
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 1. Učitavanje templatea i statičnih fajlova
	// Pazi da su putanje točne u odnosu na tvoj backend folder
	r.LoadHTMLGlob("../client/templates/*.html")
	r.Static("/assets", "../client/assets")

	// 2. Main Page - Servira tvoj index.html
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// 3. GET: Dohvaća sve monitore i renderira ih kao listu kartica (HTMX Load)
	r.GET("/api/monitors", func(c *gin.Context) {
		monitors, err := GetAllMonitors()
		if err != nil {
			c.String(http.StatusInternalServerError, "Baza ne surađuje")
			return
		}
		
		// Za HTMX, moramo "ispljunuti" sve kartice jednu za drugom
		for _, m := range monitors {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"ID":     m.ID,
				"Name":   m.Name,
				"URL":    m.URL,
				"Status": "UP", // Privremeno dok ne sredimo worker
			})
		}
	})

	// 4. POST: Dodaje novi monitor i vraća SAMO taj jedan novi box (HTMX Post)
	r.POST("/api/monitors", func(c *gin.Context) {
		// HTMX šalje "application/x-www-form-urlencoded", pa čitamo ovako:
		name := c.PostForm("name")
		url := c.PostForm("url")

		// Hvatanje ID-a i Errora iz tvoje nove SaveMonitor funkcije
		newID, err := SaveMonitor(name, url)
		if err != nil {
			c.String(http.StatusInternalServerError, "Greška pri spremanju")
			return
		}

		// VRAĆAMO SAMO JEDAN FRAGMENT (KARTICU)
		// HTMX će ovo "pljusnuti" na vrh liste pomoću hx-swap="afterbegin"
		c.HTML(http.StatusOK, "index.html", gin.H{
			"ID":     newID,
			"Name":   name,
			"URL":    url,
			"Status": "UP",
		})
	})

	// 5. DELETE: Briše monitor i vraća prazno (HTMX Delete)
	r.DELETE("/api/monitors/:id", func(c *gin.Context) {
		id := c.Param("id")
		err := DeleteMonitor(id)
		if err != nil {
			c.String(http.StatusInternalServerError, "Neuspješno brisanje")
			return
		}
		// Vraćamo 200 OK, HTMX će na frontendu maknuti element jer je hx-swap="outerHTML"
		c.Status(http.StatusOK)
	})

	return r
}