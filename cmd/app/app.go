package main

import (
	"github.com/joho/godotenv"
	"github.com/tehrelt/unreal/internal/app"
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
}

func main() {

	app, err := app.New()
	if err != nil {
		panic(err)
	}

	app.Run()

}
