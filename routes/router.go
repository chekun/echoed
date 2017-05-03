package routes

import (
	"io/ioutil"
	"strings"

	"github.com/chekun/echoed/typecho/models"

	"fmt"
	"os"

	"gopkg.in/kataras/iris.v6"
)

const ECHOED_VERSION = "1.0.0-beta1"

func InitRouters(app *iris.Framework) {
	app.Get("/", home)
	app.Get("/packages.json", packages)
	app.Get("/themes/:name/download/:version", downloadTheme)
	app.Get("/plugins/:name/download/:version", downloadPlugin)
}

func home(ctx *iris.Context) {
	ctx.Render("index.html", iris.Map{"v": ECHOED_VERSION})
}

func packages(ctx *iris.Context) {
	models.NewDb()
	f, _ := ioutil.ReadFile("storage/packages.json")
	ctx.Write(f)
}

func downloadTheme(ctx *iris.Context) {
	path, name := downloadPackage("themes", ctx)
	ctx.SendFile(path, name)
}

func downloadPlugin(ctx *iris.Context) {
	path, name := downloadPackage("plugins", ctx)
	ctx.SendFile(path, name)
}

func downloadPackage(pType string, ctx *iris.Context) (string, string) {
	path := fmt.Sprintf("%s/%s/%s/%s-%s",
		os.Getenv("PACKAGES_DIRECTORY"),
		pType,
		ctx.Param("name"),
		ctx.Param("name"),
		ctx.Param("version"))
	models.NewDb()
	models.CountVersionDownload(
		ctx.Param("name"),
		strings.Replace(ctx.Param("version"), ".zip", "", -1),
	)
	return path, ctx.Param("name") + "-" + ctx.Param("version")
}
