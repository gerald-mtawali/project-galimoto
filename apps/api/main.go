package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// album represents data about arecord album
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

var albums = []album{
	{ID: "1", Title: "Operation: Doomsday", Artist: "MF Doom", Price: 16.99},
	{ID: "2", Title: "To Pimp a Butterfly", Artist: "Kendrick Lamar", Price: 17.99},
	{ID: "3", Title: "Born Sinner", Artist: "J. Cole", Price: 19.99},
}

// getAlbums response with list of all albums as JSOn
func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

func main() {
	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.Run("localhost:8080")
}
