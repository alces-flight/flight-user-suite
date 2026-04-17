module github.com/concertim/flight-user-suite/flight-web-suite

go 1.26.1

replace github.com/concertim/flight-user-suite/flight => ../flight-core

require github.com/labstack/echo/v5 v5.1.0

require (
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	golang.org/x/time v0.14.0 // indirect
)
