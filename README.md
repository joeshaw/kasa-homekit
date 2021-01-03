
# kasa-homekit

Apple HomeKit support for TP-Link Kasa smart home devices using
[hc](https://github.com/brutella/hc).

Devices are detected and communicated with via the local network APIs.
This module does not use the cloud APIs and does not require you to log
into the Kasa cloud service.

Currently this service only supports Kasa HS1xx Smart Plugs.

Once the device is paired with your iOS Home app, you can control it
with any service that integrates with HomeKit, including Siri ("Turn
off the Christmas tree") and Apple Watch.  If you have a home hub like
an Apple TV or iPad, you can control the device remotely.

## Installing

The tool can be installed with:

    go get -u github.com/joeshaw/kasa-homekit

Then you can run the service:

    kasa-homekit

The service will search for Kasa devices on your local network at
startup, and every 5 seconds afterward.

To pair, open up your Home iOS app, click the + icon, choose "Add
Accessory" and then tap "Don't have a Code or Can't Scan?"  You should
see the Leaf under "Nearby Accessories."  Tap that and enter the PIN
00102003.  You should see one entry appear for each Kasa device on
your network.

## Contributing

This code is fairly hacky, with some hardcoded values like timeouts.
It also has limited device support.

Issues and pull requests are welcome.  When filing a PR, please make
sure the code has been run through `gofmt`.

## License

Copyright 2020 Joe Shaw

`kasa-homekit` is licensed under the MIT License.  See the LICENSE
file for details.


