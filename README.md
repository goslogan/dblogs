# DBLogs

`dblogs` is a script to generate a visual representation of a Redis Cloud account's changes over time. The input to the script is one or more downloads from the <em>Logs</em> tab of the Redis Cloud UI.

## Installation

Download from https://github.com/goslogan/dblogs/releases - binaries for all our OSes are present. If on MacOS you will need to tell the security system to allow it to run.

If you have golang installed you can do

```
git clone https://github.com/goslogan/dblogs.git
cd dblogs
go install
```
As long as you have golang installed correctly that will install to ${HOME}/go/bin which may or may not be on your PATH. 

## Running it.

Two outputs are possible - a cleansed CSV and an HTML representation. 

Parameters are as follows:

| Argument | Type | Description |
| -------- | --------- | -------- |
| -a/--accounts | comma separated list of integers |   account ids matching sources (defaults to all accounts) |
| -d/--databases | comma separated list of strings |   report only these named databases (defaults to all databases)|
| -b/--dbsort | *flag* |  sort by database name before timestamp (always true when producing html output) |
| -f/--files | comma separated list of strings | list of csv files to process (reads STDIN if not given) | 
| -F/--from string | date in yyyy-mm-dd-format |           First date to include in the output (defaults to 01 January 1900) |
| -h/--hourly | *flag* |  aggregate hourly instead of daily |
| -o/--output | filename | output file for CSV dump or HTML timeline (STDOUT by default) | 
| -s/--subscriptions |comma separated list of integers| subscription ids matching sources (defaults to all subscriptions) |
| -p/--template | path/file |   Path to a custom template for output | 
| -t/--timeline | *flag* |              generate a timeline graph for each database |
|  -i/title | string|        the title for the timeline report (default "Configuration Timeline") |
| -T/--to  | date in yyyy-mm-dd format | Last date to include in the output (defaults to 31 December 2099) |
```

Icons are used to visualise changes in the HTML output; hover over them to see the detail of the change. Icons are highlighted in green for enabled or increased configurations, red for disabled or decreased and grey for simple changes (e.g. backup path changed)>

Not all changes are represented at this time. Unparsed changes are represented by an Information icon.

Input is read from standard input if not otherwise specified and written to standard output in the same way.

## Examples

Generate a timeline HTML file with hourly aggregation

```
dblogs -f system_log.csv -o customer.html --timeline --hourly
```

Generate a cleansed CSV sorted by database.

```
cat system_log.csv | dblogs -b
```


Generate a timeline for a single subscription including changes in 2022 

```
dblogs -f system_long.csv -o customer.html --timeline --subscriptions 98765321 --from=01-01-2022 --to=31-12-2022
```
