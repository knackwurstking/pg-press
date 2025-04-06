# Notes

- JSON GET/POST/PUT/DELETE
- Auth via github oauth (?)
- Accept api tokens
- Live mode:
    - Use websockets for this mode [Scenario](#scenario)
    - Need to tell the server if someone is editing
    - Event "editing" with data `{ "user": "knackwurstking.tux@gmail.com", "data": "metal-sheets:120x60:G03" }`

```json
{
    "knackwurstking.tux@gmail.com": {
        "pg-vis": {
            "metal-sheets": {
                {
                    "format": "120x60",
                    "toolID": "G03",
                    "data": {
                        "press": 0,
                        "table": {
                            "filter": [4],
                            "data": [
                                ["13", "?", "2", "0", "", "29.0 (33.5)"],
                                ["14", "?", "1", "0", "", "30.6 (33.5)"],
                                ["8", "50", "4", "8", "", "15.8 (22.10)"]
                            ]
                        }
                    }
                }
            }
        }
    }
}
```

## Scenario

- 2 clients working on metal-sheets data for 120x60 G03
- client 1 is adding 13mm data but with data "?" for the second entry (time: 22:30:15)
- client 2 is adding 13mm data but with "47" instead of "?" (time: 22:30:14)
- each request takes 2 seconds
