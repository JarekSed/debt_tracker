package debt_tracker

import (
	"database/sql"
	"flag"
	"fmt"
    "github.com/coopernurse/gorp"
	_ "github.com/ziutek/mymysql/godrv"
	"log"
)

var databaseName = flag.String("database_name", "tcp:localhost:3306*debt_tracker", "Name of the database to connect to")
var databaseUser = flag.String("database_user", "debt_tracker", "User to connect to the database as")
var databasePassword = flag.String("database_password", "debt_tracker", "Password to connect to the database with")

// Database handles all interactions with the data model.
type Database struct {
	database *sql.DB
	dbmap *gorp.DbMap
}

// Person represents a person that can have a balance.
type Person struct {
	FirstName string
	LastName  string
    PhoneNumber uint64
}

func (p *Person) FullName() string {
	return p.FirstName + " " + p.LastName
}

func (d *Database) GetUserByPhoneNumber(phoneNumber string) (*Person, error) {
    var person Person
    err := d.dbmap.SelectOne(&person, "Select * FROM Persons where PhoneNumber=?", phoneNumber)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
	return &person, nil
}

func ConnectToDatabase() (*Database, error) {
	database_string := fmt.Sprintf("%v/%v/%v", *databaseName, *databaseUser, *databasePassword)
	db, err := sql.Open("mymysql", database_string)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
		return nil, err
	}
    err = db.Ping()
    if err != nil {
        log.Fatal("Error connecting to database:", err)
        return nil, err
    }
    // construct a gorp DbMap
    dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}

    // add a table, setting the table name to 'Persons' and
    // specifying that the id property is an auto incrementing PK
    dbmap.AddTableWithName(Person{}, "Persons").SetKeys(false, "PhoneNumber")


    // create the table. in a production system you'd generally
    // use a migration tool, or create the tables via scripts
    err = dbmap.CreateTablesIfNotExists()
    if err != nil {
        log.Fatal("Creating table failed:", err)
        return nil, err
    }

	return &Database{database: db, dbmap: dbmap}, nil
}

func (d *Database) Close() {
    d.database.Close()
}
