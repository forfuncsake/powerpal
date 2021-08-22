# Client for the Powerpal Readings API

https://readings.powerpal.net (https://www.powerpal.net)

## Device
The Powerpal is a small Bluetooth Low Energy (BLE) device that reads the flashes of the LED on a smart electricity meter and allows for a realtime view of power consumption in the home.

## App
The Powerpal mobile app, when paired with a Powerpal device:
  - displays current and historic power consumption
  - includes challenges and advice
  - forwards meter readings to the cloud; and
  - can provide a signed download link for up to 90 days of usage history

## The Missing Link
It is currently not possible for a Powerpal end user to view their usage data outside of the app without manually creating an export link for each 90 day period, then loading the provided CSV data into a visualisation tool of their choice.

This repo was created to allow for regular retrieval of the latest uploaded readings so they can be loaded into a local DB for:
  - long term storage
  - visualization on a private home dashboard
  - triggering home automation actions


## More Info

See: https://forfuncsake.github.io/post/2021/08/owning-my-own-powerpal-data/

The author of this repo is not affiliated with Powerpal.
