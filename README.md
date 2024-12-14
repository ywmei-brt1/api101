# api101

## Deployment
This repo is automatiaclly deployed to a freehost called render.
The server side logs can be seen from here: https://dashboard.render.com/web/srv-ctevjklds78s73dl7peg/logs

### To call the APIs on the server:

#### PUT (save a string on the server side, with a timestamp)
`curl -X PUT -d "hello" https://api101-2fp4.onrender.com/put`
`curl -X PUT -d "world" https://api101-2fp4.onrender.com/put`

#### GET (get back all the saved strings with their timestamps)
`curl https://api101-2fp4.onrender.com/get`

```
$ curl https://api101-2fp4.onrender.com/get
[{"timestamp":"2024-12-14T21:40:24.511207246Z","value":"hello"},{"timestamp":"2024-12-14T21:45:26.658372422Z","value":"world"}]
```

#### SEARCH (get all the saved strings match regex q)
`curl "https://api101-2fp4.onrender.com/search?q=wor"`

```
$ curl "https://api101-2fp4.onrender.com/search?q=wor"

[{"timestamp":"2024-12-14T21:45:26.658372422Z","value":"world"}]
```
