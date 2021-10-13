# solarscraper
SolarOS Scraper

Fork of https://github.com/groob/solarscraper/ to scrape data from SolarOS.

This now includes a MQTT client for communicating with [Home Assistant](https://www.home-assistant.io/). 

## Install

1. Clone repo.
2. Change to repo path.
3. Copy `config.example.toml` to `config.toml`
4. Fill in Username and Password for minimum example.
5. go run .
6. Browser should open with json data.

### Raspberry Pi daemon
1. Follow 1-4 above.
2. Fill out MQTT section in config.
3. `go build` should create a binary. In my case `v2`
4. Modify [solarscraper.service](systemd/solarscraper.service) for system and enable in systemd.

### Home Assistant
1. Enable the MQTT section in the [config file](config.example.toml).
2. To add an Energy sensor, use the [Integration integration](https://www.home-assistant.io/integrations/integration/) on the current_generation. Lifetime generation appears to be [inaccurate](#2).

```yaml
  - platform: integration
    source: sensor.current_generation
    name: Solar Panel Generated
    round: 2
```

## License

Apache-2.0 and CC-BY-SA-4.0 ([third-party files](cc-by-sa.go))
