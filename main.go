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
	IP             string
	ISOCountryCode string
	CountryName    string
	ContainerID    string
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

	resultJson, _ := json.Marshal(GeoipResult{IP: req.PathParams["ipstring"],
		ISOCountryCode: record.Country.IsoCode,
		CountryName:    record.Country.Names[0],
		ContainerID:    os.Getenv("HOSTNAME")})

	fmt.Fprint(rw, string(resultJson))
	//fmt.Printf(rw, "IP: ", req.PathParams["ipstring"], "ISO country code: %v\n", record.Country.IsoCode)

}

func (ctx *Context) LandingPage(rw web.ResponseWriter, req *web.Request) {

	fmt.Fprint(rw, "Hello from Free GeoIP Country Lookup Service that use the MaxMind GeoIP database. \n\n",
		"You can get country code for ip by doing a GET request on <this_host_url>/<ip>\n\n",
		"Example: http://geoip-service.rk44.flynnhub.com/3.3.3.3\n ",
		"..this will return a JSON result.\n\n",
		"{\"IP\":\"3.3.3.3\",\"ISOCode\":\"US\",\"ContainerID\":\"7ed8050a0106470cb9874bc681d512f1\"} \n\n\n\n",
		"powered by... \n",
		"Flynn docker PaaS - https://flynn.io\n",
		"Amazon AWS - http://aws.amazon.com\n",
		"MaxMind - http://dev.maxmind.com/geoip/geoip2/geolite2/\n\n",
		"Source code on GitHub - https://github.com/ckloppers/geoip-service\n",
		"Contact: Corné Kloppers - ckloppers@gmail.com")
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
