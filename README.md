# stdb - Strip Test Database

A database for storing data acquired during the test campaign of the LSPE/Strip
polarimeters.

This is a software that creates and handle a database of the data acquired
during the test campaign of the Strip instrument. Strip is part of the LSPE
CMB experiment.

The program has been written in Go, and the database schema it uses is based
on Sqlite3 (for test metadata) and FITS files (for test data). It provides
both a Command-Line Interface (CLI) and a Web User Interface (WebUI). Not
every operation possible from the CLI is allowed through the WebUI, but the
latter allows the database to be used remotely. WebUI requires authentication,
and it should be put behind a reverse proxy which implements TLS.

The database implements a web API to download tests; since this API does not
require authentication, it only allows read-only operations.

## Installation

Install the latest version of the Go compiler (1.8 at the time of writing) and
the following packages:

- [go-sqlite3](https://github.com/mattn/go-sqlite3) (SQLite3 driver)
- [fitsio](https://github.com/astrogo/fitsio) (FITS file I/O)
- [xls](https://github.com/extrame/xls) (Excel file parser)
- [readline](https://github.com/chzyer/readline) (CLI input)
- [cobra](https://github.com/spf13/cobra) (CLI interface)
- [viper](https://github.com/spf13/viper) (configuration files)
- [Gin](https://github.com/gin-gonic/gin) (WebUI interface)
- [UUID](https://github.com/satori/go.uuid) (managing UUIDs)
- [bcrypt](https://godoc.org/golang.org/x/crypto/bcrypt) (password hashing algorithm)

Then download this package and run

    go test

to compile it and run a test suite. If all the tests pass, you can get an
overview of `stdb`'s commands by running

    stdb help

## Example

### Command-line interface

The following commands use the CLI to create a database and add an user
and a test to it.

First, you need to create a database at some specific path; we use `./test1`. The
following command will create the directory `test1` within the current path
and initialize an empty database in it:

    $ stdb create --dbpath ./test1
    creating directory "test1"
    creating a new database file "test1/index.db"
    database "test1" created

Next, we create a new user:

    $ > stdb --dbpath ./test1 adduser --email "user@foo.bar" --fullname "Abraham Lincoln" --username "lincoln"
    Password:
    Enter again the password:
    user "lincoln" created successfully in database "test1"

Now we can add a few tests. Since this is just an example, we will use a test
data file provided in the distribution (`testdata/keithley_test.xls`):

    $ stdb --dbpath test1 add --description "A test file acquired using Keithley" \
           --polarimeter 1 --shortname "Phase switch tuning" \
           --type "PH/SW" --username "lincoln" testdata/keithley_test.xls cryo
    importing file "testdata/keithley_file.xls"
    new test with ID 1 has been created

## License

The program is released under a MIT license.