// Package weather shows current conditions and a short forecast for any
// location using Open-Meteo (no API key required). Forecasts are cached and
// refreshed in the background.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package weather
