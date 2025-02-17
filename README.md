# `pgmaster`

Find your PostgreSQL master node with a single function call.

## Why?

I had a PostgreSQL cluster with multiple nodes that would rotate the _master_
title during regular maintenance and on master node failure. I also had a few
CRON jobs that relied on getting the _read-write_ connection and they would fail
otherwise.

I needed to make sure that they find the right _master_ on init. So I wrote this
library.

## Usage

```go
func main() {
    const timeout = 5 * time.Second

    master, err := pgmaster.Find(connect, timeout, []string{
        "abc.db.example.net",
        "def.db.example.net",
    })

    // ... use master
}

// This is just an example of what it might look like.
// Use your own connection settings in practice.
func connect(host string) (*sql.DB, error) {
    const port = 5432
    return sql.Open("postgres", connectionString(host, port))
}
```

**NOTE:** the `timeout` is going to apply to each host separately and it only
really matters if some of the hosts become unavailable. For instance, with the
example timeout of 5 seconds, the longest it will take `pgmaster.Find` to check
2 hosts is 5 seconds x 2 hosts = 10 seconds. Under normal circumstances, it is
not going to take that long, because all it takes to check for a master node is
running a simple `SELECT`.
