# api101

## Deployment
This repo is automatiaclly deployed to a freehost called render.
The server side logs can be seen from here: https://dashboard.render.com/web/srv-ctevjklds78s73dl7peg/logs

### To host the code locally and call the API via local host:

See the comments in cmd/my-app/main.go

### To call the APIs (backed by the remote server):

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

#### Long Pulling (keep getting the latest change until timeout [30s])

`curl -i -w "\nExit Code: %{http_code} " https://api101-2fp4.onrender.com/get/longpoll`

```
$ curl -i -w "\nExit Code: %{http_code} " https://api101-2fp4.onrender.com/get/longpoll
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sun, 15 Dec 2024 15:36:13 GMT
Transfer-Encoding: chunked

[{"timestamp":"2024-12-15T15:36:13.907069955Z","value":"hello"}]
[{"timestamp":"2024-12-15T15:36:13.907069955Z","value":"hello"},{"timestamp":"2024-12-15T15:36:22.253719383Z","value":"hello world"}]
[{"timestamp":"2024-12-15T15:36:13.907069955Z","value":"hello"},{"timestamp":"2024-12-15T15:36:22.253719383Z","value":"hello world"},{"timestamp":"2024-12-15T15:36:28.865798362Z","value":"hello world 123"}]
Timeout

```

#### Generate QR code for google meet API:

You can copy and paste the google meet url directory from your browser then call the API:
`curl -X PUT -H "Content-Type: application/json" -d '{"link": "https://meet.google.com/oqn-ybdc-nbi?authuser=0"}' https://api101-2fp4.onrender.com/generate-qr > image.png`

or you can call this API with just the meet ID:

`curl -X PUT -H "Content-Type: application/json" -d '{"link": "https://meet.google.com/oqn-ybdc-nbi"}' https://api101-2fp4.onrender.com/generate-qr > image.png`

On linux or max, you can use `open image.png` to directly open the QR code once downloaded.

```
curl -X PUT                     \
  -H "Content-Type: application/json" \
  -d '{"link": "https://meet.google.com/oqn-ybdc-nbi?authuser=0"}' \
  https://api101-2fp4.onrender.com/generate-qr > image.png && open image.png
```