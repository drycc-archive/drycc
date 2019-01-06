# controller

This is the Drycc Controller. It is loosely inspired by the [Heroku Platform
API](https://devcenter.heroku.com/articles/platform-api-reference) and enables
management of applications running on Drycc via an HTTP API.

The controller depends on PostgreSQL and is typically booted by
[bootstrap](/bootstrap).

The API is in a state of flux and is undocumented. [cli](/cli) is one of the API
consumers.
