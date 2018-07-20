# gravo

gravo - Grafana for Volkszaehler - is an adapter for using [Grafana](https://grafana.com) with [Volkszaehler](https://volkszaehler.org).

While it is possible to run Grafana against the Volkszaehler database directly using the Grafana MySQL datasource, gravo supports additional features:

  - metrics discovery: all public channels are discoverable via the Grafana UI
  - using Volkszaehler data aggregation gravo can perform multi-year queries in sub second query time similar to the native Volkszaehler UI


## Usage

  1. have a working installation of Volkszaehler
  2. install Grafana and the [Simple JSON Datasource](https://github.com/grafana/simple-json-datasource) plugin
  3. install and run gravo

      grave -api http://myserver/middleware.php -port 8001 

  4. now create a simple json datasource and point it to gravo running on machine and port chosen before:

      http://gravo-host:8001

  5. start creating dashboards and panels
  
