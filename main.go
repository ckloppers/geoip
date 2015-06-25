package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"geoip-service/Godeps/_workspace/src/github.com/gocraft/web"
	"geoip-service/Godeps/_workspace/src/github.com/oschwald/geoip2-golang"
)

type GeoipResult struct {
	IP          string
	ISOCode     string
	ContainerID string
}

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

	resultJson, _ := json.Marshal(GeoipResult{IP: req.PathParams["ipstring"], ISOCode: record.Country.IsoCode, ContainerID: os.Getenv("HOSTNAME")})

	fmt.Fprint(rw, string(resultJson))
	//fmt.Printf(rw, "IP: ", req.PathParams["ipstring"], "ISO country code: %v\n", record.Country.IsoCode)

}

func (ctx *Context) LandingPage(rw web.ResponseWriter, req *web.Request) {

	fmt.Fprint(rw, """Hello from GeoIP-Service \n This API use the free database from MaxMind (http://dev.maxmind.com/geoip/legacy/geolite/)""")
	fmt.Fprint(rw, "You can get country code for ip by doing a GET request on host/<ip>")
}

func main() {

	router := web.New(Context{}).
		Middleware(web.LoggerMiddleware).
		Middleware((*Context).OpenMaxMindDB).
		Get("/", (*Context).LandingPage).
		Get("/:ipstring", (*Context).LookUpCountryForIp)

	http.ListenAndServe(":"+os.Getenv("PORT"), router)
	//http.ListenAndServe(":3000", router)
}
