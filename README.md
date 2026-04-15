# RSS Aggregator (for BootDev)

This is a CLI-based RSS aggregator done for a BootDev course.

## Requirements

You will need to have **PostgreSQL** and **Go** installed in order to user this.

## Installation

You can install the latest version with

`go install github.com/KJBrock/bootdev_gator@latest`

## Configuration

The configuration file is located in ~/.gatorconfig.json

The initial configuration should be 

```
{
  "db_url" : <postgres connection>
}
```

Where the PostgreSQL connection string will be something like

`yourusername@localhost:5432/gator?sslmode=disable`

