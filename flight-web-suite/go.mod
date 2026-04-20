module github.com/concertim/flight-user-suite/flight-web-suite

go 1.26.1

replace github.com/concertim/flight-user-suite/flight => ../flight-core

require (
	github.com/PuerkitoBio/goquery v1.10.3
	github.com/gorilla/sessions v1.4.0
	github.com/labstack/echo/v5 v5.1.0
	golang.org/x/net v0.52.0
)

require (
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	golang.org/x/text v0.36.0 // indirect
	golang.org/x/time v0.14.0 // indirect
)
