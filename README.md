### How to

##### Start

- `make install` to install the dependencies (lint, migration) [optional]
- `docker compose up` to start postgres locally
- apply env variables (depends on your IDE) from `local.env` file
- [optional] override MIGEATRIONS_DIR folder since my IDE run is from "projectDir/cmd/server"
- start the app `go run cmd/server/main.go`

##### Test

- `make test_unit` to run unit tests
- `make test_integration` to run integration tests

### Code base layout

##### Transport layer

`handlers` folder contains transport layer of the applicaiton. It's responsible for serving external requests and providing the data from the application domain layer.
It includes:
- reading the incoming requests
- writing the outgoing responses
- handling domain layer errors and represent them to the clients
- catching unexpected behaviour and log such events

##### Domain layer

`domain` folder contains domain layer of the app. It's responsible for:
- data validation
- data retreiving from the given data source
- necessary calculations around the given data

It's important to note all the models are defined as values, not pointers.
The pointers should be used only if we don't want to copy the object, it's useful for holding database connections, mutexes (https://gobyexample.com/mutexes) and any other state.
It will not improve the application performance since the runtime will spend more time on looking it at the heap for accessin and GCing.

##### Data layer

`repository` folder contains data layer of the ap. It's responsible for 
- retreiving data from the database
- mapping the database rows to the domain data models
- handling data errors and present them as domain errors
in a nutshell it must implement the interface the domain layer expects, not the other way around.

##### assembly

It's a folder responsible for composing all the dependencies and providing the core components for the process such as web service, logger, migration launcher and so on.

##### pkg

The folder keeps all the internal dependencies. They can potentially be moved to another repo/package to serve more applications. In our particular example we keep a logger there, middleware to log the requests and inject a logger into a context instance for attaching the given request (trace-id) to all the logged messages.

### The other parts of the codebase

##### Tests specifics

`testdata` folder is responsible to hold the fixtures. 
It's a specific name to let the compiler know it can ignore it: "The go tool will ignore a directory named "testdata", making it available
to hold ancillary data needed by the tests."
type `go help test` to read more.

##### Mocks

A regular package has a dependency. The depence is defined as an interface providing the described api.
In order to isolate the tests the mocks are provided.
Such interfaces have a `go:generate` comment to execute a defined command during `go generate` command.
Every time an interface is updated you must to run `make gen` in order to update the mocks. 
The mocks are located in the package subdirectory `mocks` and used only in the directory test files.

##### Database schema

The users table is very simple. However, a few details I want look closer.
The `deletedAt` column is there might look as antipattern. GDPR makes it more complicated and sometimes we need a background job to catch "soft-deleted" rows, collect the archive, send to a defined direction and then completely remove the data saving the anonimyzed part of it for analytics or others goals.

There are also columns such as `updatedAt` and `createdAt` that are never exposed to API until the requirement is writtend down.

### Points of improvements

##### AuthN and AuthZ
There are 2 main ways to implement it:
- static token (jwt as a good example)
- sessions

First, AuthN.
Sessions are good at event driven applications such as messagers where every event(message) should be delivered not to a user, but every device/session.
Every session must:
- have its own event series
- be able to revoke the others sessions
- be able to sign out itself not imopacting the others
- issue a refresh token only for the current session

For most of the applications a static token is suitable enough.
The app issues a pair of tokens, access and refresh.
Access is a short lived token used to authorize (AuthZ) the user.
After expiration a client must use a refresh token in order to request a new pair of tokens.

A simple and reliable way to implement AuthZ is RBAC (role based access control).
Every user role has it's  own way to create an account, the applicaiton grants a designated role.
As a result the applicaiton Authorization middleware decides whether to accept or reject a user's request.

For more complicated domains sometimes ABAC (attributes based access control) required. Where a limited amount of roles isn't enough the app can grant specific attributes to a user and match them.

In order to simplify the delivery there are plenty solutions on the market:
- keycloak
- ory
- zitadel
etc.

##### Pagination

The pagination is implemented in the simplest way, getting `limit` and `offset` inputs.
The pages related data is queries in the same query in order to keep the result consistent.
We can separate those queries either fetching the data in the same transaction or being ready to get inconsistent data.


There are a few more options available:
- accepting `page` and `size` instead, it removed the calculation `offset` calculation from a client side with very little disadvantage.
- cursor pagination, it makes it strongly coupled to a database
- scrolling, it's an option to provide additional chunk of the content without explicit paging, every next chunk is requested based on the last/first item attributes (timestamp + ID since timestamp itself is not a unique value)

##### Migrations

It's not the best practice to apply migrations on the applicaiton start.
It's good to have an init container or a job definition that will make a backup first and then apply the migrations.

##### Integration tests

The integration tests must be separated by API.
Currently they are not and the potential failure can take more time to narrow the point of failure.
It takes a couple more steps to prepare adding a database fixtures (especially for read-only API).
Also, it's good to compare the test results to embedded json fixtures instead of domain models. Having an issue in the marshalling or the model definition will not detect the issue.

##### Test coverage
Since Go1.20 it supports coverage for the integration tests as well:
https://go.dev/blog/integration-test-coverage

##### Database schema
The datatabase schema should be extended to handle contact data, authentication methods (if many or password hash as another solution), and sign in identity unique constraint.

##### Api spec
The api spec (openapi/swagger) must be defined for the service.
We can write this document manually.
It's good to integrate lazy tools such as http://huma.rocks in order to generate the api spec and never write it manually (register fake web server and expose only openapi page from it).

### Scalability
The application itself has no much resource intencive load, it does IO operations providing access to the database. Therefore, I anticipate most of the bottleneck on the database size, there are many well known practices we can apply in order to serve more users/requests:
- define the most complex queries and create a specific index for it (for example, on `createdAt` if that's the most frequent data range we query); currently Postgres doesn't support unique keys on hash indexes, therefore even uuid Primary Key has a btree index, so we could add hash index on the `id` field, as a result the reading a single item operation performance might be slightly improved
- setup cache instances (lru,lfu) such as redis, it's even better to setup for multi-region application in order to put the data closer to the users, as a result even more reducing latency
- add more database replicas, it's important not to use statement-based replica (should be deprecated for most of the database) since we use non-deterministic queries such as calling `now()` funciton in the database
- sharding the database, the easiest way to shard by region; if the app is write-heavy we could create partitions based on createdAt timestamp

After all, creating an application replicas could add more network throughput
