# Routes

## HTML

| Method | Path                         | Description                  |
| ------ | ---------------------------- | ---------------------------- |
| GET    | /                            | Home page                    |
| GET    | /login                       | Login page                   |
| GET    | /logout                      | Logout and redirect to login |
| GET    | /feed                        | Feed page                    |
| GET    | /profile                     | User profile page            |
| GET    | /trouble-reports             | Trouble reports page         |
| GET    | /trouble-reports/share-pdf   | Share PDF export             |
| GET    | /trouble-reports/attachments | Get attachment data          |
| GET    | /tools                       | Tools page                   |
| GET    | /tools/active/:press         | Active tools by press        |
| GET    | /tools/all/:id               | Tool details by ID           |

## HTMX

| Method | Path                                      | Description                          |
| ------ | ----------------------------------------- | ------------------------------------ |
| GET    | /htmx/nav/feed-counter                    | WebSocket for feed counter updates   |
| GET    | /htmx/feed/data                           | Get feeds table                      |
| GET    | /htmx/profile/cookies                     | Get cookies table                    |
| DELETE | /htmx/profile/cookies                     | Delete cookie from table             |
| GET    | /htmx/trouble-reports/dialog-edit         | Edit dialog                          |
| POST   | /htmx/trouble-reports/dialog-edit         | Submit new trouble report            |
| PUT    | /htmx/trouble-reports/dialog-edit         | Update existing trouble report       |
| GET    | /htmx/trouble-reports/data                | Render trouble reports list          |
| DELETE | /htmx/trouble-reports/data                | Delete a trouble report              |
| GET    | /htmx/trouble-reports/attachments-preview | Get attachments preview              |
| GET    | /htmx/trouble-reports/modifications/:id   | Render trouble report modifications  |
| POST   | /htmx/trouble-reports/modifications/:id   | Reset trouble report to modification |
| GET    | /htmx/tools/edit                          | Tools edit dialog                    |
