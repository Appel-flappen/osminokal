# Osminokal

Osminokal is your little tentacled sea-dwelling friend to help you connect
Octopus Energy's free energy sessions to your calendar.

## Usage

Osminokal is configured by environment variables only. Here is an example:

```bash
OSMINOKAL_SOURCE="whizzy"
OSMINOKAL_CALENDARS='[{"name":"mycaldav", "type":"Caldav", "endpoint":"https://mydomain.tld/dav/calendars/username/default", "username":"user", "password":"pass"}]'
OSMINOKAL_LOG_LEVEL="info"
OSMINOKAL_POLL_INTERVAL="15m"
OSMINOKAL_CA_CERT_PATH="/home/user/certs/root_ca.crt"
OSMINOKAL_CLIENT_CERT_PATH="/home/user/certs/client.crt"
OSMINOKAL_CLIENT_CERT_KEY_PATH="/home/user/certs/client.key"
```

You can specify your own certs if your calendar is behind mTLS.

The source can be either `"whizzy"` or `"david"` (not implemented yet). These
are two different providers of the octopus free energy session information.
Octopus does not expose an API for this yet, so these kind people scrape their
emails and provide a json api for us to use.

> [!WARNING]
> Keep the interval nice and high to avoid overloading these APIs provided generously to us.
> The sessions are always announced around 24 hrs in advance, you have plenty of time to get the events, no need to spam.

Calendars are the place to send the events. You can create as many calendars as
you want. They are configured with a json list in the env var
`OSMINOKAL_CALENDARS`. Check below for detailed keys depending on the calendar
connector you are using.

Then simply run osminokal! It will startup and poll on the defined interval for new events.

## Configuration

### Calendars

#### CalDAV

For the CalDAV connector, you need to have the following keys:

```json
[
  {
    "name": "mycaldav",
    "type": "Caldav",
    "endpoint": "https://mydomain.tld/dav/calendars/username/default",
    "username": "user",
    "password": "pass"
  }
]
```

- `name` is just for your reference
- `type` must be caldav (case insensitive) to select the caldav connector.
- `endpoint` is expected to be the full link to the calendar you want the events sent to.
- `username` is your username for plain auth
- `pasword` is your password for plain auth

# todos

- [ ] calendars need deps, but source doesnt?
- [ ] Implement david source
