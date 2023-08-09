# DBConfig

A script to generate a visual representation of a Redis Cloud account's changes over time. The input to the script is one or more downloads from the <em>Logs</em> tab of the Redis Cloud UI.

Two outputs are possible - a cleansed CSV and an HTML representation. 

Parameters are as follows:

```
  -a, --accounts uints        account ids matching sources (default [0])
  -d, --databases strings     report only these named databases
  -b, --dbsort                sort by database name before timestamp
  -f, --file strings          list of csv files to process
  -h, --hourly                aggregate hourly instead of daily
  -o, --output string         output file for CSV dump or HTML timeline
  -s, --subscriptions uints   subscription ids matching sources (default [0])
  -t, --timeline              generate a timeline graph for each database
```

Icons are used to visualise changes in the HTML output; hover over them to see the detail of the change. Icons are highlighted in green for enabled or increased configurations, red for disabled or decreased and grey for simple changes (e.g. backup path changed)>

Not all changes are represented at this time. Unparsed changes are represented by an Information icon.

Input is read from standard input if not otherwise specified and written to standard output in the same way.

## Examples

Generate a timeline HTML file with hourly aggregation

```
dbconfig -f system_log.csv -o customer.html --timeline --hourly
```

Generate a cleansed CSV sorted by database.

```
cat system_log.csv | dbconfig -b
```