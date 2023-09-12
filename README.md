# Autoheater

## Description

A CLI to automatically turn on/off your heater/cooler based on prices and weather to save money

## Motivation

Domotic devices are cool, but they will be cooler when [Matter](https://csa-iot.org/all-solutions/matter/) 
protocol lands, finally, into the industry. 

As a fact, before Matter, only some protocols such as Zigbee had some standardization mercy. 

But there are a lot of devices out there connected by Wi-Fi, and they are completely non-standard. Manufacturers 
decided to design them without thinking about potential integration between different stuff.

Nowadays, this situation improved a lot with [HASS scripts](https://www.home-assistant.io/integrations/script/), 
but still not everyone wants to maintain a HASS instance just to turn on the heater (or the cooler) and anything else.

Moreover, the prices are always changing, so automating this with fixed hours will not keep your electricity bill low.
In addition, doing complex automation scripts is simply forcing you to have advanced systems like Home Assistant, 
which is really, really wonderful, but sometimes complex

This CLI comes to solve those issues. But how?

* Simple CLI (or Docker image) with closed configuration that automates the whole process
* Takes into account daily prices for electricity, and weather in your zone to find the cheapest hours for heating or cooling
* Turn on/off your device on the cheapest moments on its own (standalone), or trigger an event to external systems 
  using little standard integrations (webhooks, MQTT, etc)

> MQTT will be included in following releases. Keep in touch to know more. 
> Do you want another integration? let's discuss it: open an issue

and doing automation scripts is simply forcing you to have advanced systems like Home Assistant, which is wonderful

## Flags

As every configuration parameter can be defined in the config file, there are only few flags that can be defined.
They are described in the following table:

| Name          | Description                        |      Default      | Example                      |
|:--------------|:-----------------------------------|:-----------------:|:-----------------------------|
| `--config`    | Define the path to the config file | `autoheater.yaml` | `--config ./autoheater.yaml` |
| `--log-level` | Define the verbosity of the logs   |      `info`       | `--log-level info`           |

## Examples

Here you have a complete example. More up-to-date one will always be maintained in 
`config/samples` directory [here](./config/samples)

```yaml
apiVersion: v1alpha1
kind: Autoheater
metadata:
  name: laundry-room-heater
spec:

  global:
    # Main scheduler calculates the schedules for the day just at 00:00h. But when the application is started
    # on a different moment, commonly in the middle of the day, may be, the best N hours were already passed.
    # This option is allowing to select the next cheapest N hours in the first startup even if they are
    # more expensive than the real cheapest ones
    ignorePassedHours: true

  # Take into account the weather as first filter. The idea is not to switch the heater on really hot days
  weather:
    enabled: true

    coordinates:
      latitude: 28.1562300
      longitude: -16.6359200

    #
    temperature:
      # Type of temperature to take into account. Possible values: apparent or real
      # Attention: apparent is recommended as it is the perceived feels-like temperature combining
      # wind chill factor, relative humidity and solar radiation
      type: apparent

      # Possible values are: fahrenheit or celsius
      unit: celsius

      # Max temperature to switch the heater on. Switching on the heater will be ignored on higher temperatures
      threshold: 30

  # Prices for today's day are coming from Apaga Luz, as these data are already filtered and ease-to-access
  # Ref: https://raw.githubusercontent.com/jorgeatgu/apaga-luz/main/public/data/today_price.json
  # Ref: https://raw.githubusercontent.com/jorgeatgu/apaga-luz/main/public/data/canary_price.json
  price:
    # Spanish pricing zone due to geographical differences. Possible values: mainland or canaryislands
    zone: canaryislands

  # Configuration related to the device
  device:

    # The type of the device to act on. This is used together with 'weather.temperature.threshold'.
    # In case 'heater' is selected, temperatures higher than the threshold won't act
    # In case 'cooler' is selected, temperatures lower than the threshold won't act
    # Possible values: cooler, heater
    type: heater

    # Time to keep the device turned on.
    # At this moment, the cheapest N hours are always the chosen ones
    activeHours: 6

    # Several integrations are covered to use this CLI as 'standalone' process, or as a possible adaptor
    # between different domotic systems (sending the events to an HTTP endpoint, mqtt, etc.)
    # ATTENTION: All configured integrations will act at the same time
    integrations:

      # Data for sending the events to TAPO P1XX devices (p100, p110, etc)
      tapoSmartPlug:
        address: "192.168.1.100"
        auth:
          username: placeholder@gmail.com
          password: 'xxxPLACEHOLDERxxx'

      # Endpoints to send the request on events
      # POST <url>: { event: 'start', name: 'pepito', timestamp: ''}
      webhook:
        url: "https://webhook.site/a7303a4b-4377-49d7-b109-6106fbe21052"
        # (Optional) username and password for basic auth
        auth:
          username: 'placeholder'
          password: 'placeholder'
```

> ATTENTION:
> If you detect some mistake on the config, open an issue to fix it. This way we all will benefit

## How to deploy

This project provides binary files and Docker images to make it easy to be deployed wherever wanted

### Binaries

Binary files for most popular platforms will be added to the [releases](https://github.com/achetronic/autoheater/releases)

### Kubernetes

An entire example is placed inside [deploy](./deploy) directory.

> ⚠️ Disclaimer: Take care of hardening it, as passwords are not read from environment variables yet. 
> Are you willing to contribute?

### Docker

Docker images can be found in GitHub's [packages](https://github.com/achetronic/autoheater/pkgs/container/autoheater) 
related to this repository

> Do you need it in a different container registry? I think this is not needed, but if I'm wrong, please, let's discuss 
> it in the best place for that: an issue

## How to contribute

We are open to external collaborations for this project: improvements, bugfixes, whatever.

For doing it, open an issue to discuss the need of the changes, then:

- Fork the repository
- Make your changes to the code
- Open a PR and wait for review

The code will be reviewed and tested (always)

> We are developers and hate bad code. For that reason we ask you the highest quality
> on each line of code to improve this project on each iteration.
