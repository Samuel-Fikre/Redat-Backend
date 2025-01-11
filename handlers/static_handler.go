package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func ServeMapUI(c *fiber.Ctx) error {
	return c.SendFile("./static/index.html")
} 