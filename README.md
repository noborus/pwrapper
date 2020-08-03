# PWrapper

Wrap and execute command with pipe connected

## Usage

```sh
pwrapper --wrap-command "psql -At" --start "BEGIN;" --end "COMMIT;"  --exec "./batch.sh"
```