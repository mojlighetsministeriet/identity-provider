[![Build Status](https://travis-ci.org/mojlighetsministeriet/identity-provider.svg?branch=master)](https://travis-ci.org/mojlighetsministeriet/identity-provider)

[![Coverage Status](https://coveralls.io/repos/github/mojlighetsministeriet/identity-provider/badge.svg?branch=master)](https://coveralls.io/github/mojlighetsministeriet/identity-provider?branch=master)

# identity-provider

Handles the management of accounts (properties id, email, roles and passwords) and JWT tokens. The service is meant to run with Docker and will need an external database service to persist the accounts to.

**NOTE: This service is still in it's experimentall phase,** feel free to try it out, contribute pull requests but for now, expect the functionality to be somewhat limited. E.g. right now there is no way of creating the first administrator account yet. It will be added but we are not there yet. :)

## Docker image

Our docker image is avaliable here https://hub.docker.com/r/mojlighetsministeriet/identity-provider/.

## Try it out

Replace user, password, host, databasename below with credentials for a running MySQL service and run the command.

    $ docker run --name identity-provider -p 1323:1323 -e DATABASE="*user*:*password*@*host*/*databasename*?charset=utf8mb4,utf8&parseTime=True&loc=Local" mojlighetsministeriet/identity-provider

For production, make sure to set RSA_PRIVATE_KEY environment variable externally to keep active tokens valid when starting new containers of this image. If you skip this a new key will be generated each time a container is created (and it will not be able to read any previous client tokens).  

## Creating a private RSA key

If you do not have a favorite tool yet, use the tool we provide https://github.com/mojlighetsministeriet/rsa-private-key-generator, see the GitHub page for installtion/usage.

## Configuration

The service is configured by environment variables. Below is a description of the ones that you need to care about in a production environment.

### RSA_PRIVATE_KEY

An RSA private key, if this variable is not set, a new key will be generated and saved to key.private, keep in mind that this key would be lost whenever the docker container is removed which would then invalidate any generated tokens and because of that require the client to re-authenticate.

### DATABASE_TYPE

The supported types are mysql, postgres, mssql. The default value is mysql.

### DATABASE

The database connection string, the service uses the package GORM and to see details about the connection string format (depending on database type) see http://jinzhu.me/gorm/database.html.

## The API structure

To start exploring the API, visit the root (depending on where you are running e.g. http://localhost:1323). This will give you a list back of all registered end points.

There are two resources, token (use for authentication) and account (used to persist who gets to create tokens and which roles each account has)

### Authenticate

To authenticate you need to make sure that you have an email and password for an account that is inside the service. Then make the following call:

    POST { "email": "user@example.com", "password": "thesupersecretpassword" } http://localhost:1323/token

If the credentials where correct you will recieve a response in the following format:

    { "token": "alongsecretjwttoken" }

Make sure to always pass this token in the request headers to any service that is connected to this service like so (more details see https://jwt.io/introduction/#how-do-json-web-tokens-work-):

    Authorization: Bearer <alongsecretjwttoken>

### Renewal

The token will expire after some time so if it's used in a UI, make sure to renew the token every now and then by:

    POST http://localhost:1323/token/renew

If the token in the headers was still valid you will recieve a response in the following format:

    { "token": "alongsecretjwttoken" }

If you call for renewal after the token has expired (as of writing set to 20 minutes) the client will have to re-authenticate instead.

### Create an account

Make sure that the client has a valid token from a previous account with the **administrator** role and call

    POST { "email": "user@example.com", "password": "thesupersecretpassword", "roles": ["user", "administrator"] } http://localhost:1323/account

### List accounts

Make sure that the client has a valid token from a previous account with the **administrator** role and call

    GET http://localhost:1323/account

## License

All the code in this project goes under the license GPLv3 (https://www.gnu.org/licenses/gpl-3.0.en.html).
