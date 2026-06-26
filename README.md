# Micronova2MQTT

Micronova2MQTT is a bridge between Micronova Agua IOT pellet-stove controllers and MQTT-based home-automation systems.
It lets you monitor and control your stove from Home Assistant, Domoticz, Node‑RED, or other MQTT systems.
Because it uses the Micronova API — which supports controllers from multiple brands — Micronova2MQTT is compatible with a wide range of pellet heating systems.

Supported [brands](brands.yml):

* Alfaplam
* Amg
* Boreal
* Bronpi
* Cola
* Corisit
* Elcofire
* Elfire
* Eva Calòr
* Fontana Forni
* Fonte Flamme
* Globe Fire
* Jolly Mec
* Karmek One
* Klover
* Laminox
* La Nordica Extraflame
* Linea VZ
* Lorflam
* MCZ (Easy Connect, Turbofonte, Easy Connect Plus & Easy Connect Poêle apps)
* Moretti
* Micronova
* Nobis
* Nordic Fire
* Piazzetta
* Ravelli
* Solartecnik EOSS
* Stufe
* Thermoflux
* Tim Sistem
* Unical

## Key Features

### Intelligent Configuration & Setup
* Automatic UUID creation and registration
* Simplified configuration — only specify the Brand; customer code and API URL are sourced from brands.yml
* RegKey translation - Micronova RegKeys can be converted to meaningful MQTT topics

### Performance & Session Management
* Smart token handling with automatic storage and refresh
* Persistent storing of Product ID and Device ID minimizes API calls across sessions
* Minimal API usage — calls are limited to once per hour during device inactivity

### Security
* Encrypted session data protecting sensitive tokens
* Configurable custom MQTT payload values to switch the pallet stove On/Off - a non-standard 'on' or 'off' value adds extra security

## Installation

Docker compose example:

```yml
services:
  Micronova2MQTT:
    image: legobas/micronova2mqtt:latest
    container_name: micronova2mqtt
    environment:
      - LOGLEVEL=info
      - TZ=America/New_York
    volumes:
      - /home/legobas/micronova2mqtt:/data:rw
    restart: unless-stopped
```

## Configuration

The settings of Micronova2Mqtt are defined by the `micronova2mqtt.yml` yaml configuration file.

## Example micronova2mqtt.yml Configuration file

```yml
mqtt:
    url: mqttbroker:1883
    username: test
    password: pass
micronova:
    brand: alfaplam
    email: user@mail.com
    password: 'SecretP@ssw'
```

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

## Environment variables

The logging level can be defined by environment variable LOGLEVEL:

```
LOGLEVEL = INFO (default)
LOGLEVEL = DEBUG
LOGLEVEL = ERROR
```

## Security

### On/Off values.

The default values to switch the pellet stove are `on` and `off`.
Switching the pellet stove On or Off can be done by sending the MQTT messages:

    micronova2mqtt/set/Power = on
    micronova2mqtt/set/Power = off

To make these values less obvious they can be obfuscated by setting the config settings:

    micronova:
        power:
            on:  secret1
            off: secret2

These on/off values can be used by sending the MQTT messages:

    micronova2mqtt/set/Power = secret1
    micronova2mqtt/set/Power = secret2

### Session storage

The session data is stored in the file `session.dat`.
This file is encrypted because it contains sensitive data like the JWT tokens.

## The Brands file

To use micronova2mqtt with a new Pellet Stove brand copy the [brands](brands.yml) file to your data directory and add your Pellet Stove brand. The app-name, customer-code and domain URL have to be provided.
If this works for you please create a pull request so other owners of the same brand can benefit from it.

## Inspired by:

* [home_assistant_micronova_agua_iot](https://github.com/vincentwolsink/home_assistant_micronova_agua_iot)
* [ioBroker.micronova](https://github.com/TA2k/ioBroker.micronova)
