# Refactoring models & services

Old directories (packages):

    -  .models/
    -  .services/
    -  .utils/

- The models and utilities are now inside the shared package "internal/shared/"
- New services are in "internal/services/"
- Service helper functions for sprcific tasks are in "internal/helper/"
- URL related helper functions are now in "internal/urlb

## What this project is about?

- I use this for storing and keep track of data
- There are 5 presses for ceramic tiles production
- We have two types of presses, one would be SITI and the other is SACMI
- There are to types of tools which fits into these presses, one will be mounted on top and the other one on bottom
- Each tool has its own unique "code" and "type", the type is just the type the tool is from (ex.: MASS, GTC, FC, ...)
- The upper tools have one addigional slot (optional) for a so called cassette
- The cassette is used for different tile thicknesses

## What is still missing?

- The service for the trouble reports

## What needs to be done?

- The 'internal/handlers/' contains all echo router handlers and templates and needs to be changed to make it work with the new services
- Shared a-h templ components are now inside the 'internal/components/'
- The modification system needs to be kicked from this project
- The resolved model paramerter needs to be replaced, i don't want these models anymore
