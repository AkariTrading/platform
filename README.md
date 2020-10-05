# Platform

## Auth Routes

### POST /auth/login

Request
```
{
  "email": "",
  "password": "",
}
```

### POST /auth/register

Request
```
{
  "email": "",
  "password": "",
}
```

### POST /auth/logout


## Api - Script Routes
### GET /api/scripts

Response
```
[
    {
        "id": "",
        "userId": "",
        "createdAt": "0001-01-01T00:00:00Z",
        "updatedAt": "0001-01-01T00:00:00Z",
        "title": ""
    }
]
```

### GET /api/scripts/{scriptID}

Response
```
{
    "id": "",
    "userId": "",
    "createdAt": "0001-01-01T00:00:00Z",
    "updatedAt": "0001-01-01T00:00:00Z",
    "title": ""
}
```

### PUT /api/scripts/{scriptID}

Request
```
{
    "title": ""
}
```

### DELETE /api/scripts/{scriptID}

### GET /api/scripts/{scriptID}/versions

Response
```
{
	"id": "",
	"scriptId": "",
	"createdAt": "0001-01-01T00:00:00Z",
	"body": ""
}
```

### POST /api/scripts/{scriptID}/versions

Request
```
{
	"body": ""
}
```



## Api - Symbol History Routes
### GET /api/history/{exchange}/{symbol}

Response
```
{
	"start": 0, // unix timestamp of the first trade for symbol
	"end": 0, // last timestamp
	"candles": [[0.0, 0.0, 0.0, 0.0, 0.0]]
}
```

## Api - Jobs Routes
### GET /api/jobs/{jobID}

Response
```
{
	"id": "",
	"exchangeId": "",
	"isRunning": false,
	"userId": "",
	"body": "",
	"createdAt": "0001-01-01T00:00:00Z",
	"lastAliveAt": "0001-01-01T00:00:00Z",
	"stoppedAt": null,
	"type": "",
	"exchange": "",
	"symbolA": "",
	"symbolB": "",
	"stateJSON": ""
}
```

### DELETE /api/jobs/{jobID}

### POST /api/jobs/{jobID}

Request
```
{
	"body": "",
	"exchange": "",
	"exchangeId": "",
	"symbolA": "",
	"symbolB": "",
	"type": "cycle", // or cycledryrun
	"balance": {"BTC": 0, "USD": 1000,
	"state": null,
	"userID": "",
}
```

### GET /api/jobs/{jobID}/logs?createdBefore={timestamp}

Response
```
[
    {
        "createdAt": "0001-01-01T00:00:00Z",
        "body": ""
    }
]
```

### GET /api/userExchanges/

Response
```
[
    {
        "id": "",
        "userID": "",
        "exchange": "",
        "createdAt": "0001-01-01T00:00:00Z"
    }
]
```

### POST /api/userExchanges/

Request
```
[
    {
        "Exchange": "",
        "ApiKey": "",
        "ApiSecret": ""
    }
]
```

Response 
```
{
    "id": "",
    "userID": "",
    "exchange": "",
    "createdAt": "0001-01-01T00:00:00Z"
}
```

### DELETE /api/userExchanges/{exchangeID}