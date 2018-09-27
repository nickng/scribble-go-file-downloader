# httpget
Scribble-Go file downloader demo.

This demos the Scribble-Go framework with the Downloader protocol
found in `Downloader.scr`.

## Usage

    $ go generate # generates the API for the protocol.
    $ go build main.go
    $ ./httpget http://example.com/

## Notes

The downloader uses one Fetcher by default.
To change the number of Fetchers used to n:

    $ ./httpget -N n

See `httpget -h` for more details about the usage.
