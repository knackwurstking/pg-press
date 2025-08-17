# Routes

@TODO: Continue here... List all routes here

## HTML

| Method | Path                         | Description                   |
| ------ | ---------------------------- | ----------------------------- |
| GET    | /                            | Home page                     |
| GET    | /login                       | Login page                    |
| GET    | /logout                      | Logout and redirecto to login |
| GET    | /feed                        | Feed page                     |
| GET    | /profile                     | User profile page             |
| GET    | /tools                       | Tools page                    |
| GET    | /tools/active/:press         | Active Tools page             |
| GET    | /tools/all/:id               | Tool page                     |
| GET    | /trouble-reports             | Trouble Reports page          |
| GET    | /trouble-reports/share-pdf   | Share PDF                     |
| GET    | /trouble-reports/attachments | Get attachment data (image)   |

## HTMX

| Method | Path                                                | Description                                     |
| ------ | --------------------------------------------------- | ----------------------------------------------- |
| GET    | /htmx/nav/feed-counter                              | Render the feed counter (span for the nav item) |
| GET    | /htmx/feed/data                                     | Get feeds table                                 |
| GET    | /htmx/profile/cookies                               | Get cookies table                               |
| DELETE | /htmx/profile/cookies?value=                        | Delete cookie from table                        |
| GET    | /htmx/trouble-reports/dialog-edit?id=&cancel=       | Edit dialog                                     |
| POST   | /htmx/trouble-reports/dialog-edit                   | Submit new data, close dialog or error          |
| PUT    | /htmx/trouble-reports/dialog-edit?id=               | Update existing data, close dialog or error     |
| GET    | /htmx/trouble-reports/data                          | Render trouble reports list                     |
| DELETE | /htmx/trouble-reports/data                          | Delete a trouble report from list               |
| GET    | /htmx/trouble-reports/attachments-preview?id=&time= | Get the attachments preview                     |
| GET    | /htmx/trouble-reports/modifications/:id             | Render trouble report modifications list        |
| POST   | /htmx/trouble-reports/modifications/:id?time=       | Reset trouble report to modified time           |
| GET    | /htmx/tools/list-all                                | List all tools                                  |
