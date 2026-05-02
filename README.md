# micronova2mqtt

Micronova2MQTT is a bridge between Micronova Agua IOT pellet-stove controllers and MQTT-based home-automation systems
It lets you monitor and control your stove from Home Assistant, Domoticz, Node‑RED, or other MQTT systems.
Because it uses the Micronova API — which supports controllers from multiple brands — Micronova2MQTT is compatible with a wide range of pellet heating systems.

Supported [brands](brands.yml):

* Alfaplam
* Amg
* Boreal
* Bronpi
* Cola
* Corisit
* Elfire
* Eva-calor
* Fontana-forni
* Fonte-flamme
* Globe-fire
* Jolly-mec
* Karmek-one
* Klover
* Laminox
* Lorflam
* MCZ (Easy Connect, Easy Connect Plus & Easy Connect Poêle apps)
* Moretti
* Micronova
* Nobis
* Nordic-fire
* Piazzetta
* Ravelli
* Solartecnik-eoss
* Stufe
* Thermoflux
* Tim-sistem
* Linea-vz

## Key Features

* Automatic UUID creation and registration
* Only the Brand has to be specified in config, customer code and API URL are chosen from brands.yml
* RegKey translation: convert Micronova RegKeys to meaningful MQTT topics
* Tokens are stored and refreshed if expired
* Product ID and Device ID are remembered between sessions to reduce API calls
* The API is called only once an hour if the device is inactive

## Brands file

To add a new Pellet Stove brand copy the [brands](brands.yml) file to your data directory and add your Pellet Stove brand. You have to know the app-name, customer-code and domain URL.
If this works for you please create a pull request so other owners of the same brand can benefit from it.

## Inspired by:

* [home_assistant_micronova_agua_iot](https://github.com/vincentwolsink/home_assistant_micronova_agua_iot)
* [ioBroker.micronova](https://github.com/TA2k/ioBroker.micronova)
