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
* Simplified configuration — only specify the Brand; the customer code and API URL are sourced from brands.yml
* Reduced RegKeys set - option to use only relevant Micronova RegKeys
* RegKey translation - change sometimes strange Italian RegKey names to customized meaningful titles
* Actions - Set specific values if a threshold is reached

### Performance & Session Management
* Smart token handling with automatic storage and refresh
* Persistent storage of Product ID and Device ID minimizes API calls across sessions
* Minimal API usage — calls are limited to once per 90 minutes during device inactivity

### Security
* Encrypted session data protecting sensitive tokens
* Configurable custom MQTT payload values to switch the pellet stove On/Off - a non-standard 'on' or 'off' value adds extra security

## Quick start

1. Create a directory for the configuration and session data:

   ```bash
   mkdir -p /home/legobas/micronova2mqtt
   ```

2. Create `/home/legobas/micronova2mqtt/micronova2mqtt.yml` with your MQTT broker details and Micronova account credentials:

   ```yaml
   mqtt:
     url: mqttbroker:1883
     username: your_mqtt_username
     password: your_mqtt_password

   micronova:
     brand: your_pellet_stove_brand
     email: you@example.com
     password: 'your_micronova_password'
   ```

3. Save this as `compose.yml`:

   ```yaml
   services:
     micronova2mqtt:
       image: legobas/micronova2mqtt:latest
       container_name: micronova2mqtt
       environment:
         - LOGLEVEL=info
         - TZ=Europe/London
       volumes:
         - /home/legobas/micronova2mqtt:/data:rw
       restart: unless-stopped
   ```

   Start the container from the directory containing `compose.yml`:

   ```bash
   docker compose up -d
   ```

4. Check startup logs:

   ```bash
   docker compose logs -f micronova2mqtt
   ```

   Look for successful MQTT and Micronova connections. To list available RegKeys, set `log_reg_keys: true` under `micronova` in the config, then restart the container.

5. Confirm MQTT messages are arriving. Subscribe to the configured base topic (default: `micronova2mqtt`) using an MQTT client or broker console. You can also send a Power command:

   ```bash
   mosquitto_sub -h mqttbroker -t 'micronova2mqtt/#' -v
   ```

   ```bash
   mosquitto_pub -h mqttbroker -t 'micronova2mqtt/set/Power' -m 'on'
   ```

   Replace `mqttbroker` with your broker’s hostname. Use `on`/`off` unless you configured custom Power values.
   
## Configuration

These are all possible options for the `micronova2mqtt.yml` yaml configuration file:

```markdown
- `mqtt` — MQTT connection settings
  - `url` — MQTT server URL (**required**)
  - `username` / `password` — MQTT server credentials; may be omitted (default: empty)
  - `qos` — MQTT Quality of Service (default: `0`, `AtMostOnce`)
  - `retain` — Retain MQTT messages (default: `false`)
  - `base_topic` — Base topic for Micronova2MQTT messages (default: `micronova2mqtt`)

- `micronova` — Micronova account and device settings
  - `brand` — Pellet stove brand or app (**required**)
  - `email` — User email address (**required**)
  - `password` — User password (**required**)
  - `power` — Power-switch payload values
    - `on` — Payload for the `On` switch (default: `on`)
    - `off` — Payload for the `Off` switch (default: `off`)
  - `log_reg_keys` — Log all available RegKeys at startup (default: `false`)
  - `actions` — Actions to run when a threshold is reached
    - `trigger` — Trigger condition
      - `get_key` — `****_get` RegKey to read
      - `min_value` — Minimum threshold value
    - `set_values` — Values to set when the action runs
      - `set_key` — `****_set` RegKey to write
      - `value` — Target value
  - `reg_keys` — RegKeys to include, with optional customized titles
    - `key` — Parameter RegKey
    - `title` — Rename or translate the parameter’s JSON field name
```

## Environment variables

The logging level can be defined by environment variable `LOGLEVEL`:

```
LOGLEVEL = info (default)
LOGLEVEL = debug
LOGLEVEL = error
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

## Optimizations

* If the device is active (the pellet stove is burning), the current values will be read every 20 seconds.
If the device is not active the period between read actions will be 90 minutes (don't hammer the API).
After a MQTT set operation the parameters are updated immediately.
* Using only necessary RegKeys can significantly reduce memory usage and MQTT/network traffic.

## Actions

**Automated Startup Optimization**

Pellet stove manufacturers often provide specific startup recommendations, some stoves have ignition sequences that should be followed exactly.
This process can be automated using the Actions feature to reduce power once the chimney reaches optimal operating temperature. This eliminates manual monitoring during the critical startup phase.

### Manual Startup Process

Before configuring automation, understand the recommended steps:

1. **Start at medium-high power** to establish draft safely
2. **Monitor chimney temperature** as the stove reaches operating temperature
3. **Reduce to desired power level** once the chimney reaches **190°C**

### Automating with Actions

This startup sequence can be fully automated through the Actions configuration. 
On ignition, the stove will operate at high power, then automatically reduce its output when the flue-gas temperature reaches 190°C.

```yaml
micronova:
    actions:
        - trigger:
              get_key: status_get
              min_value: 1
          set_values:
              - set_key: power_set
                value: 4
              - set_key: temp_air_set
                value: 21
              - set_key: vent_main_set
                value: 3
              - set_key: canalization_1_set
                value: 2
        - trigger:
              get_key: temp_gas_flue_get
              min_value: 190
          set_values:
              - set_key: power_set
                value: 1
              - set_key: vent_main_set
                value: 1
              - set_key: canalization_1_set
                value: 1
```

**Note:** Always refer to your stove's manual for manufacturer-specific startup recommendations and ignition sequences, as procedures vary by model.

## RegKey minimization and translation

Using only the necessary RegKeys can significantly reduce memory usage and MQTT/network traffic.
RegKeys can also be given clearer, more meaningful names — for example, replacing Italian names such as giri_estrattore_get and ore_lavoro_par_get with descriptive English equivalents.
If you use actions, include their RegKeys in the reduced parameter list; otherwise, the actions won’t work!

```yml
micronova:
    reg_keys:
        - key: status_get
          title: Status
        - key: alarms_get
          title: Alarm
        - key: giri_estrattore_get
          title: Extractor Rotation Speed
        - key: ore_lavoro_par_get
          title: Operating Hours
```

## How do I know which RegKeys are available?

If the config parameter `log_reg_keys` is set to `true` all available RegKeys will be written to the log once after the service is started.

```yml
micronova:
    log_reg_keys: true
```

Check the log:

```bash
docker compose logs micronova2mqtt | grep RegKey:
```

Example output:

```
INFO   RegKey: alarms_enable
INFO   RegKey: alarms_get
...
```

## Extended example micronova2mqtt.yml Configuration file

```yml
mqtt:
    url: mqttbroker:1883
    username: test
    password: pass
micronova:
    brand: alfaplam
    email: user@mail.com
    password: 'SecretP@ssw'
    power:
        on: secret1
        off: secret2
    log_reg_keys: true
    actions:
        - trigger:
              get_key: temp_gas_flue_get
              min_value: 190
          set_values:
              - set_key: power_set
                value: 1
              - set_key: vent_main_set
                value: 1
              - set_key: canalization_1_set
                value: 1
    reg_keys:
        - key: status_get
          title: Status
        - key: alarms_get
          title: Alarm
        - key: power_set
          title: SetPower
        - key: temp_air_set
          title: Thermostat
        - key: temp_air_get
          title: TempRoom
        - key: temp_gas_flue_get
          title: TempFlueGas
        - key: vent_main_set
          title: SetVentilationSpeed
        - key: vent_front_get
          title: Ventilation
```

## Building and running

Build with:

    go build -ldflags "-X main.Version=1.0.0"

The `micronova2mqtt.yml` file has to exist in one of the following locations:

 * A `data` directory in the filesystem root: `/data/micronova2mqtt.yml` (used for the docker image)
 * A `.data` directory in the user home directory `~/.data/micronova2mqtt.yml`
 * The current working directory
 * A `data` directory in the current working directory

## The Brands file

To use Micronova2MQTT with a new pellet stove brand copy the [brands](brands.yml) file to your data directory and add your brand.
The app-name, customer-code and domain URL have to be provided.
If this works for you please create a pull request so other owners of the same brand can benefit from it.

## Inspired by:

* [home_assistant_micronova_agua_iot](https://github.com/vincentwolsink/home_assistant_micronova_agua_iot)
* [ioBroker.micronova](https://github.com/TA2k/ioBroker.micronova)
