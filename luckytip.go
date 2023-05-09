package main

import (
	"strconv"

	"github.com/antoniomralmeida/luckytip/megasena"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

func main() {
	MS, _ := megasena.CreateFactory()
	app := fiber.New(fiber.Config{
		Views: html.New("./views", ".html")})

	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("main", nil)
	})
	app.Post("/", func(c *fiber.Ctx) error {
		value, _ := strconv.ParseFloat(c.FormValue("valor"), 64)
		bets, _ := MS.Aposta(value)
		return c.Render("main", bets)
	})

	app.Listen(":8080")

}
