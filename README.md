[![Build Status](https://travis-ci.org/mojlighetsministeriet/identity-provider.svg?branch=master)](https://travis-ci.org/mojlighetsministeriet/identity-provider)
[![Coverage Status](https://coveralls.io/repos/github/mojlighetsministeriet/identity-provider/badge.svg?branch=master)](https://coveralls.io/github/mojlighetsministeriet/identity-provider?branch=master)

# identity-provider

Handles the management of accounts (properties id, email, roles and passwords) and JWT tokens. The service is meant to run with Docker and will need an external database service to persist the accounts to.

## Try it out

Replace user, password, databasename below with credentials for a running MySQL service and run the command.

    $ docker run --name identity-provider -p 1323:1323 -e DATABASE="*user*:*password*@/*databasename*?charset=utf8mb4,utf8&parseTime=True&loc=Local" mojlighetsministeriet/identity-provider

For production, make sure to set RSA_PRIVATE_KEY environment variable externally to keep active tokens valid when starting new containers of this image. If you skip this a new key will be generated each time a container is created (and it will not be able to read any previous client tokens).  

## Help with creating a private RSA key

Use the tool https://github.com/mojlighetsministeriet/rsa-private-key-generator, see the GitHub page for installtion/usage.

## Configuration

The service is configured by environment variables. Below is a description of the ones that you need to care about in a production environment.

### RSA_PRIVATE_KEY

An RSA private key, if this variable is not set, a new key will be generated and saved to key.private, keep in mind that this key would be lost whenever the docker container is removed which would then invalidate any generated tokens and because of that require the client to re-authenticate.

### DATABASE_TYPE

The supported types are mysql, postgres, mssql. The default value is mysql.

### DATABASE

The database connection string, the service uses the package GORM and to see details about the connection string format (depending on database type) see http://jinzhu.me/gorm/database.html.

## License

All the code in this project goes under the license GPLv3 (https://www.gnu.org/licenses/gpl-3.0.en.html).
