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
* Only the Brand has to be specified in the config file, customer code and API URL are provided by the brands.yml file
* RegKey translation: convert Micronova RegKeys to meaningful MQTT topics
* Tokens are stored and refreshed if expired
* Product ID and Device ID are remembered between sessions to reduce API calls
* The API is called only once an hour if the pellet stove is inactive

## Config

The settings of Micronova2Mqtt are taken from the `micronova2mqtt.yml` yaml configuration file.
The `micronova2mqtt.yml` file has to exist in one of the following locations:

 * A `data` directory in de filesystem root: `/data/micronova2mqtt.yml`
 * A `.data` directory in the user home directory `~/.data/micronova2mqtt.yml`
 * The current working directory
 * A `data` directory in the current working directory

## Configuration options

| Config item               | Description                                                              |
| ------------------------- | ------------------------------------------------------------------------ |
| **mqtt**                  |                                                                          |
| $~~$ url                  | MQTT Server URL                                                          |
| $~~$ username/password    | MQTT Server Credentials (can be omitted)                                 |
| $~~$ qos                  | MQTT Server Quality Of Service                                           |
| $~~$ retain               | MQTT Server Retain messages                                              |
| **micronova**             |                                                                          |
| $~~$ brand                | Pellet stove brand / app                                                 |
| $~~$ email                | User email address                                                       |
| $~~$ password             | User password                                                            |
| $~~$ **power**            | on/off secrets                                                           |
| $~~~~$ on                 | Secret for the `On` switch                                               |
| $~~~~$ off                | Secret for the `Off` switch                                              |
| $~~$ **reg_keys**         | RegKey translations                                                      |
| $~~~~$ key                | Parameter RegKey                                                         |
| $~~~~$ topic              | Parameter Topic/Title                                                    |


## Brands file

To add a new Pellet Stove brand copy the [brands](brands.yml) file to your data directory and add your Pellet Stove brand. You have to know the app-name, customer-code and domain URL.
If this works for you please create a pull request so other owners of the same brand can benefit from it.

## Inspired by:

* [home_assistant_micronova_agua_iot](https://github.com/vincentwolsink/home_assistant_micronova_agua_iot)
* [ioBroker.micronova](https://github.com/TA2k/ioBroker.micronova)
