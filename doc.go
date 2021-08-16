// Package powerpal provides a basic client for retrieving meter readings from
// the Powerpal readings API for a given device with its authorization token.
// API: https://readings.powerpal.net/documentation
package powerpal

/*
Sample Queries:


GET https://readings.powerpal.net/api/v1/device/<serial> HTTP/2.0
accept:	application/json
accept-encoding:	gzip, deflate, br
user-agent:	Powerpal/1895 CFNetwork/1240.0.4 Darwin/20.5.0
accept-language:	en-au
authorization:	<token>


Resp:
{
    "available_days": 8,
    "first_reading_timestamp": 1624954980,
    "last_reading_cost": 0.003244248,
    "last_reading_timestamp": 1625743740,
    "last_reading_watt_hours": 22,
    "serial_number": "<serial>",
    "total_cost": 40.96497,
    "total_meter_reading_count": 13147,
    "total_watt_hours": 215354
}


GET https://readings.powerpal.net/api/v1/meter_reading/<serial>?start=1624954980&end=1625743740 HTTP/2.0
accept:	application/json
content-type:	application/json
accept-encoding:	gzip, deflate, br
user-agent:	Powerpal/1895 CFNetwork/1240.0.4 Darwin/20.5.0
authorization:	<token>
accept-language:	en-au

Resp (gzip encoded):
[
    {
        "cost": 0.002801850,
        "is_peak": false,
        "pulses": 60,
        "samples": 1,
        "timestamp": 1624954980,
        "watt_hours": 19
    },
    {
        "cost": 0.002801850,
        "is_peak": false,
        "pulses": 61,
        "samples": 1,
        "timestamp": 1624955040,
        "watt_hours": 19
    },
    {
        "cost": 0.002801850,
        "is_peak": false,
        "pulses": 61,
        "samples": 1,
        "timestamp": 1624955100,
        "watt_hours": 19
    },
	...,
    {
        "cost": 0.003244248,
        "is_peak": false,
        "pulses": 69,
        "samples": 1,
        "timestamp": 1625743740,
        "watt_hours": 22
    }
]
*/
