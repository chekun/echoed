package main

import (
	"log"
	"os"

	"github.com/chekun/echoed/routes"
	"github.com/joho/godotenv"
	"github.com/kataras/iris/adaptors/httprouter"
	"github.com/kataras/iris/adaptors/view"
	"gopkg.in/kataras/iris.v6"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := iris.New()
	app.Adapt(
		iris.DevLogger(),
		httprouter.New(),
		view.HTML("./views", ".html").Reload(os.Getenv("DEBUG") == "true"),
	)

	routes.InitRouters(app)

	app.Listen(":6300")

}
