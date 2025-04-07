# Notes

> Sever address: "knackwurstking.duckdns.org/pg-vis"

- Telegram bot integration, a button where the user can request an api key,
  just use Alice for this
- ~Websockets for realtime editing, need to support multiple users here~
- Websockets for server events, or just use these sse events
- PWA, for offline support
- Vite Frontend makes the pwa stuff more easy, maybe think about using svelte-kit
  for this project
- Add a app download button for easy install, so noob users can handle this
- Everything will be synced with the server for each user using the api key
  for identification
- Need some kind of an admin page bound to as api key (the master key)
- Replace the gist id page with the api key page, also preview permissions the
  client gets with his api key

## Questions

- Should this be a MPA or a SPA (the main app /pg-vis)
- How to send valid emails via golang
- How to package a multipage PWA golang web app with capacitor for a good
  android app with proper routing
- Do i want an capacitor android app here?
- Find out how to do pwa auto updates faster

## What i need

- Golang
    - Server framework or just use the stdlib
        - [**echo**](https://echo.labstack.com/docs/quick-start)
    - A local database, maybe somethings like sqlite(?)
        - [**bbolt**](https://github.com/etcd-io/bbolt) (_key/value store_)

## New Categories

- Howtos (multi user edit)
- Problem Reports, moved out of alert lists category (multi user edit)
    - Allow voting for deletion, but only the admin is allowed to do the deletion
      or someone with special permissions
- Handbooks, mostly pdfs (multi user edit, special permission required)

## Routes

| Path                 | Description                                                       |
| -------------------- | ----------------------------------------------------------------- |
| /pg-vis              | Start page: Api key setup, Api key permissions, Web app news, ... |
| /pg-vis/registration | Register API key and info about how to get one                    |
