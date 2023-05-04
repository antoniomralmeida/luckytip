package main

import (
	"fmt"

	"github.com/antoniomralmeida/luckytip/megasena"
	"github.com/gofiber/fiber/v2"
)

func main() {
	MS, _ := megasena.CreateFactory()
	app := fiber.New()

	fmt.Println(MS.Aposta(70.3))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(MS)
	})
	//app.Listen(":8080")

}
