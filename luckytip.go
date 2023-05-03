package main

import (
	"fmt"

	"github.com/antoniomralmeida/luckytip/megasena"
)

func main() {
	MS, err := megasena.CreateFactory()
	fmt.Println(MS, err)

}
