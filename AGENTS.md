# PG Press Handover Document

## Project Overview

- This project uses golang, (a-h) templ for html templating, Echo for routing,
  and a custom ui library for styling. and htmx for dynamic content loading.
- Styling is mostly done using the ui library inside the assets directory
  "internal/assets/assets/css/ui.min.css"
- Generating .templ file using the `templ` tool, (ex.: `templ generate`) or
  just use `make generate`

## Project Structure

### Templating

- Shared components are stored in "internal/components"
- Page-specific templates are stored in "internal/handlers/{handler_name}/templates/"
- In handlers, each route should be inside a separate file

## Git Commit Message Conventions

- Always use semantic git commit message conventions.
