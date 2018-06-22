package api

import (
	"log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net"
	"fmt"
	"github.com/urfave/cli"
	"os"
	"bufio"
	"strings"
	"net/http"
)

type Ret struct{
	Postcode string
	Addr string
	Status int
}

func Run(c *cli.Context) {

	e := echo.New()

	// public
	e.Static("/public", "public")

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	data, err := loadData(c.String("source"))
	if err != nil {
		e.Logger.Fatal(err)
	}

	e.GET("/", func(ctx echo.Context) (err error) {
		code := ctx.QueryParam("code")
		r := &Ret{
			Postcode: code,
			Status: 1,
		}
		if addr, ok := data[code]; ok {
			r.Addr = addr
			r.Status = 0
		}
		return ctx.JSON(http.StatusOK, r)
	})

	// Start server
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", c.String("bind"), c.String("port")))
	if err != nil {
		e.Logger.Fatal(l)
	}

	e.Listener = l
	e.Logger.Fatal(e.Start(""))

	if c.Bool("debug") {
		log.Printf("http server started on %s:%s", c.String("bind"), c.String("port"))
	}
}

func loadData(path string) (data map[string]string, err error){
	data = make(map[string]string)
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		//log.Println(scanner.Text())
		arr := strings.Split(scanner.Text(), ",")
		if len(arr) == 4 {
			data[arr[0]] = fmt.Sprintf("%s%s%s", arr[1], arr[2], arr[3])
		}
	}
	log.Println("load data success")
	return
}