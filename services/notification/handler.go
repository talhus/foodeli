package notification

import (
	"foodeli/platform/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup, authMW gin.HandlerFunc) {
	r.GET("/notifications/stream", authMW, h.Stream)
}

func (h *Handler) Stream(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	//headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Flush()

	ch := h.hub.Register(userID)
	defer h.hub.Unregister(userID, ch)

	c.SSEvent("connected", "connected")
	c.Writer.Flush()
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			c.SSEvent("message", string(msg))
			c.Writer.Flush()
		}
	}

}
