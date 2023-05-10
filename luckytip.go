package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/antoniomralmeida/luckytip/megasena"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

func main() {
	var port = "8080"
	if len(os.Args) == 1 {
		fmt.Println("Use luckytip -p <PORT> to define, default is 8080")
	} else {
		if os.Args[1] == "-p" {
			port = os.Args[2]
		}
	}
	MS, err := megasena.CreateFactory()
	MS.CreateBarChart()
	if err != nil {
		log.Fatal(err)
	}
	app := fiber.New(fiber.Config{
		Views: html.New("./views", ".html")})

	app.Use(logger.New())

	app.Static("/", "./views")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("main", nil)
	})
	app.Post("/", func(c *fiber.Ctx) error {
		value, _ := strconv.ParseFloat(c.FormValue("valor"), 64)
		bets, _ := MS.Aposta(value)
		return c.Render("main", bets)
	})
	log.Fatal(app.Listen(":" + port))
}
