package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"geoip-service/Godeps/_workspace/src/github.com/gocraft/web"
	"geoip-service/Godeps/_workspace/src/github.com/oschwald/geoip2-golang"
)

var port string

type Context struct {
	db *geoip2.Reader
}

func (ctx *Context) OpenMaxMindDB(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	db, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx.db = db
	next(rw, req)
}

func (ctx *Context) LookUpCountryForIp(rw web.ResponseWriter, req *web.Request) {

	// If you are using strings that may be invalid, check that ip is not nil
	ip := net.ParseIP(req.PathParams["ipstring"])
	record, err := ctx.db.Country(ip)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf(rw, "IP: ", req.PathParams["ipstring"], "ISO country code: %v\n", record.Country.IsoCode)
	fmt.Fprint(rw, "IP: ", req.PathParams["ipstring"], "  ISO country code: ", record.Country.IsoCode)

}

func (ctx *Context) LandingPage(rw web.ResponseWriter, req *web.Request) {

	fmt.Fprint(rw, "Hello from Flynn on port " ,port " from container ", os.Getenv("HOSTNAME"), "\nYou can get country code for ip by doing a GET request on ", os.Getenv("HOSTNAME"),"/<ip>")
}

func main() {

	port := os.Getenv("PORT")

	router := web.New(Context{}).
		Middleware(web.LoggerMiddleware).
		Middleware((*Context).OpenMaxMindDB).
		Get("/", (*Context).LandingPage).
		Get("/:ipstring", (*Context).LookUpCountryForIp)

	http.ListenAndServe(":"+port, router)
}
