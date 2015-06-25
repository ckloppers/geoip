package main

//
// Corne Kloppers
// 25 June 2015
// Quick hack job to read MaxMind GeoIP database and get a ISO Country code based on IP
// This also follow a patern to be able to deployed on AWS using Flynn. Heroku like PaaS
//
// Hack alert - Don't email me if things don't work :)
// ckloppers@gmail.com
//

// Go imports
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

// JSON strcut format
type GeoipResult struct {
	IP             string
	ISOCountryCode string
	ContainerID    string
}

// Context used within App
type Context struct {
	db *geoip2.Reader
}

// Read MaxMind database file
func (ctx *Context) OpenMaxMindDB(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	db, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx.db = db
	next(rw, req)
}

// Main working function. Parse IP from request object and read DB to get ISOCode for country and contruct JSON result
func (ctx *Context) LookUpCountryForIp(rw web.ResponseWriter, req *web.Request) {

	// If you are using strings that may be invalid, check that ip is not nil
	ip := net.ParseIP(req.PathParams["ipstring"])
	record, err := ctx.db.Country(ip)
	if err != nil {
		log.Fatal(err)
	}

	// Mashall result JSON with data in it
	resultJson, _ := json.Marshal(GeoipResult{IP: req.PathParams["ipstring"],
		ISOCountryCode: record.Country.IsoCode,
		ContainerID:    os.Getenv("HOSTNAME")})

	fmt.Fprint(rw, string(resultJson))
	//fmt.Printf(rw, "IP: ", req.PathParams["ipstring"], "ISO country code: %v\n", record.Country.IsoCode)

}

// Fancy landing page :-)
func (ctx *Context) LandingPage(rw web.ResponseWriter, req *web.Request) {

	fmt.Fprint(rw, "Hello, from the \"Free\" GeoIP Country Lookup Service that uses the MaxMind GeoIP lite database. \n\n",
		"You can get the ISO Country Code for a specified IP, by doing a GET request on this URL <this_host_url>/<ip>\n\n",
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

// Main entry
func main() {

	router := web.New(Context{}).
		Middleware(web.LoggerMiddleware).
		Middleware((*Context).OpenMaxMindDB).
		Get("/", (*Context).LandingPage).
		Get("/:ipstring", (*Context).LookUpCountryForIp)

	http.ListenAndServe(":"+os.Getenv("PORT"), router)
	//http.ListenAndServe(":3000", router)
}
