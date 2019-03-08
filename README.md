# gravo
[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=BB3W3WH7GVSNW)

gravo - Grafana for Volkszaehler - is an adapter for using [Grafana](https://grafana.com) with [Volkszaehler](https://volkszaehler.org).

While it is possible to run Grafana against the Volkszaehler database directly using the Grafana MySQL datasource, gravo supports additional features:

- metrics discovery: all public channels are discoverable via the Grafana UI, private channels can also be used by adding them via their UUID
- custom channel name: you can change the channels name for Grafana
- performance: using Volkszaehler data aggregation gravo can achieve sub-second query times even when retrieving multiple years of data similar to the native Volkszaehler UI


## Installation

### Prerequisites

  1. Install [Volkszaehler](https://github.com/volkszaehler/volkszaehler.org)

  2. Install Grafana and the [JSON Datasource](https://github.com/simPod/grafana-json-datasource) plugin. [Simple JSON Datasource](https://github.com/grafana/simple-json-datasource) will also work but not allow you to specify additional query parameters.

### gravo

gravo can either be run using docker or built manually:

    docker run -p 8000:8000 andig/gravo -api http://myserver/middleware.php

Building and running gravo manually:

    make
    gravo -api http://myserver/middleware.php -url 0.0.0.0:8000 

### Grafana datasource

Create a Grafana Simple JSON Datasource and point it to gravo running on machine and port chosen before:

    http://gravo-host:8000

## Usage

To use gravo for querying Volkszaehler data. Create Grafana panels for gravo datasource and add metrics:

- metric can use the channel name if the channel is public
- alternatively the UUID of a private channel can be used

### Customization

Using the [JSON Datasource](https://github.com/simPod/grafana-json-datasource), the Volkszaehler query can further be tailored by adding "Additional JSON Data":

- To **override the UUID with the channel name** add:

      {"name": "channel title"}

- Volkszaehler can **generate aggregated/averaged data** by time period:

      {"group": "hour/day/month"}

- To **improve Volkszaehler response times** gravo is able to optimize queries. In order to do so the number of expected result tuples can be specified. If not specified Volkszaehler will return data in highest resultion which can potentially be millions of records:

      {"tuples": 500}

  This requires active data aggreation in the Volkszaehler installation.

- Volkszaehler can also **return "raw", uninterpreted data** which is e.g. useful for retrieving meter reading values or meters pulses without pulse to power conversion. Use the following to return daily consumption values:

      {"options": "raw"}

- `"options"` can also be used to **send other options** that the Volkszaehler middleware accepts. One example is retrieving consumption data per period:

      {
          "group": "day",
          "options": "consumption"
      }

  Note: consumption data requires volkszaehler next (andig/volkszaehler.org)
  
### Example

Below is an example of a complex Grafana dashboard for Volksaehler:

  ![Panel](https://github.com/andig/gravo/blob/master/doc/dashboard.png)

See [json](https://github.com/andig/gravo/blob/master/doc/dashboard.json) for the example dashboard source.