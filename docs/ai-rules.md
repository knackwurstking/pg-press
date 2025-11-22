# PG Press

This project uses golang, a-h templ, echo and sqlite for the database

## Directory Structure

| Path                           | Description                                            |
| ------------------------------ | ------------------------------------------------------ |
| /services                      | Database handlers/services                             |
| /handlers                      | Echo (web) handlers                                    |
| /handlers/components           | Shared (a-h) templ components between handlers         |
| /handlers/components/oob       | OOB Components and rendering utilities                 |
| /handlers/<handler>/components | Components only use from inside this handler goes here |
| /handlers/dialogs              | Dialog handler and somponents                          |
| /models                        | Models                                                 |
| /pdf                           | pdf generating functions                               |
| /utils                         | utility functions                                      |
| /errors                        | All the error related stuff here                       |
| /env                           | Environment variables/constants here                   |
| /docs                          | All Documentation stuff here                           |
| /assets                        | All assets the echo server will serve at ""            |

## General

- JavaScript inside ".templ" files should never ever use `const` or `var`.
- For html styling always use the the `/assets/css/ui.min.css` library if possible, keep it minimal if not possible to style with this library.
