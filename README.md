# mixport


`mixport` utilizes Mixpanel's
[data export API](https://mixpanel.com/docs/api-documentation/exporting-raw-data-you-inserted-into-mixpanel)
to download event data and make it far easier to consume and store.

Mixpanel exposes this data as a bunch of JSON blobs in the form `{"event":
"EventName", "properties": {...}}`, where the `properties` map contains all of
the arbitrary dimensions attached to the data point. There isn't a defined
schema, property keys and event types can come and go, data types can
change. Such is JSON.

This does however make storing the data in a SQL database somewhat of a chore,
since you can really only try to guess the valid types (I wrote and maintained
this application as the precursor to `mixport`, believe me, this isn't fun or
very logical).

`mixport` is really only concerned with the export portion of the problem, but
can do some streaming transforms on the data to make it more digestible.

## Usage

First, let's set up a configuration file:

```bash
$ cp mixport.conf{.example,}
$ $EDITOR mixport.conf
```

There are several comments in the example configuration which should give a
decent explanation of what's required and available to configure.

Let's start off by just downloading yesterday's data for all the products
specified in the configuration:

```bash
$ ./mixport
```

That's it! If your configuration isn't in `.` or is named something other than
`mixport.conf`, you need to specify it with `-c path/to/config`.

If you want to download data for a single day (that isn't yesterday), use the
`-d` flag:

```bash
$ ./mixport -d 2013/12/31
```

If you want to download multiple days of data, try this:

```bash
$ ./mixport -r 2013/12/31-2014/02/06
```

And that's about all you need to know to get started.

## Export formats

There are currently 3 exportable formats:

### Schemaless CSV

This format is meant to ultimately be COPY friendly for SQL databases and gets
around the problems mentioned above about data types and column inconsistency
by being schemaless. The output is simply: `event_id,key,value`

So for example, a data point input of `{"event": "Foo", "properties": {"bar":
"baz", ...}}` becomes this:

```CSV
event_id,key,value
some_UUID,event,Foo
some_UUID,bar,baz
...
```

If you can insert this data into a SQL table where the `event_id`, `key`, and
`value` columns are all `VARCHAR` (or equivalent). You can then `GROUP BY` the
`event_id` to get a row.

This subverts a bunch of advantages to SQL, so make sure you know this is the
best way.

Note that this exploded format does get pretty big on disk, and hugely benefits
from GZIP compression.

### Flattened JSON

So as I've just explained, the input is in the form `{"event": "Foo",
"properties": {"bar": "baz", ...}}`.

This export basically folds the 'properties' map up into the top level:

`{"event": "Foo", "bar": "baz", ... }`

It's possible to preserve the original hierarchy by writing a new (or modifying
this) export writer, but I don't really see why you would want to.

Like the CSV export, this is compressed by GZIP very efficiently. 85-90%
compression ratio is typical for data I've looked at.

### Amazon Kinesis

Simple enough, this repeatedly calls
[`PutRecord`](http://docs.aws.amazon.com/kinesis/latest/APIReference/API_PutRecord.html)
with the data points in the flatten JSON format. Kinesis application to process
data is left as an exercise left to the reader.

**Warning:** haven't gotten around to testing this one yet.

It's really simple to write a new one (seriously, just look at the
[JSON one](https://github.com/boredomist/mixport/blob/master/exports/json.go)),
so if you have an idea for an export, please contribute.

## Mixpanel to X without hitting disk

`mixport` can write to [named pipes](http://en.wikipedia.org/wiki/Named_pipe)
instead of regular files, which can be incredibly useful to avoid temporary
files.

The basic formula is to set `fifo=true` in the `[json]` or `[csv]` sections of
the configuration file and then just pretend that the named pipe has all of the
data already written to it, like so:

```bash
$ ./mixport
$ for fifo in /mixport/output/dir/*; do psql -c "COPY TABLE ... FROM $fifo" & done
```

If you want to pipe things to S3 without hitting disk, I've had success with
the [s3stream gem](https://github.com/kindkid/s3stream):

```bash
$ for fifo in /mixport/output/dir/*; do s3stream mybucket $fifo <$fifo & done
```

Neat.
