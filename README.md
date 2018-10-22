# gravo

gravo - Grafana for Volkszaehler - is an adapter for using [Grafana](https://grafana.com) with [Volkszaehler](https://volkszaehler.org).

While it is possible to run Grafana against the Volkszaehler database directly using the Grafana MySQL datasource, gravo supports additional features:

  - metrics discovery: all public channels are discoverable via the Grafana UI
  - performance: using Volkszaehler data aggregation gravo can achieve sub-second query times even when retrieving multiple years of data similar to the native Volkszaehler UI


## Usage

  1. have a working installation of [Volkszaehler](https://github.com/volkszaehler/volkszaehler.org)
  2. install Grafana and the [JSON Datasource](https://github.com/simPod/grafana-json-datasource) plugin. [Simple JSON Datasource](https://github.com/grafana/simple-json-datasource) will also work but not allow you to specify additional query parameters.
  3. install and run gravo

          gravo -api http://myserver/middleware.php -port 8001 

  4. now create a simple json datasource and point it to gravo running on machine and port chosen before:

          http://gravo-host:8001

      Example:

      ![Datasource](https://github.com/andig/gravo/blob/master/doc/datasource.png)

  5. start creating dashboards and panels

      Example:

       ![Panel](https://github.com/andig/gravo/blob/master/doc/panel.png)

## Building

To build for your platform:

    go build -o gravo *.go
