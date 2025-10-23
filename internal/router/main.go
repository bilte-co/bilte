package router

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	healthcheck "github.com/RaMin0/gin-health-check"
	brotli "github.com/anargu/gin-brotli"
	"github.com/bilte-co/bilte/internal/domain"
	"github.com/bilte-co/bilte/internal/templates"
	"github.com/gin-gonic/gin"

	// "github.com/nats-io/nats.go"
	stats "github.com/semihalev/gin-stats"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(50, 200)

func rateLimiter(c *gin.Context) {
	if !limiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
		c.Abort()
		return
	}
	c.Next()
}

func HeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}

func NewRouter(r *gin.Engine, production *bool) *gin.Engine {
	r.Use(brotli.Brotli(brotli.DefaultCompression))
	r.Use(healthcheck.Default())
	r.Use(stats.RequestStats())
	r.Use(rateLimiter)

	r.GET("/stats", func(c *gin.Context) {
		c.JSON(http.StatusOK, stats.Report())
	})

	r.GET("/", func(c *gin.Context) {
		title := "bilte co"
		description := "Strategy-led software engineering and consulting—delivering impactful results in high-stakes domains."
		c.HTML(http.StatusOK, "", templates.Home(production, &title, &description))
	})

	r.GET("/cv", func(c *gin.Context) {
		var Info domain.Resume
		var Projects domain.Projects
		// we are going to get the data in data/resume.json and data/projects.json
		infoFile, err := os.Open("data/resume.json")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error opening info file: "+err.Error())
			return
		}
		defer infoFile.Close()

		infoBytes, err := io.ReadAll(infoFile)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error reading info file: "+err.Error())
			return
		}
		err = json.Unmarshal(infoBytes, &Info)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error unmarshalling info file: "+err.Error())
			return
		}

		projectsFile, err := os.Open("data/projects.json")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error opening projects file: "+err.Error())
			return
		}
		defer projectsFile.Close()

		projectsBytes, err := io.ReadAll(projectsFile)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error reading projects file: "+err.Error())
			return
		}

		err = json.Unmarshal(projectsBytes, &Projects)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error unmarshalling projects file: "+err.Error())
			return
		}

		c.HTML(http.StatusOK, "", templates.CV(production, Info, Projects))
	})

	return r
}
