# solarscraper
SolarOS Scraper

Fork of https://github.com/groob/solarscraper/ to scrape data from SolarOS.

This now includes a MQTT client for communicating with HomeAssistant. Please enable the MQTT section in the [config file](config.example.toml).

## Install

1. Clone repo.
2. Change to repo path.
3. Copy `config.example.toml` to `config.toml`
4. Fill in Username and Password for minimum example.
5. go run .
6. Browser should open with json data.

## License

Apache-2.0 and CC-BY-SA-4.0 ([third-party files](cc-by-sa.go))