package main

import (
	"github.com/antoniomralmeida/luckytip/megasena"
	"github.com/gofiber/fiber/v2"
)

func main() {
	MS, _ := megasena.CreateFactory()
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(MS)
	})
	app.Listen(":8080")

}
