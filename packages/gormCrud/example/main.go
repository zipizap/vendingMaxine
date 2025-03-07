package main

import (
	"example/gormCrud"
	"fmt"
	"os"

	"encoding/json"

	"github.com/google/uuid"
)

type User struct {
	gormCrud.GormCrud[User]
	Name   string
	Email  string
	Age    uint
	Active bool
	// 1User-to-1Profile relationship
	Profile *Profile
	// 1User-to-manyAddress relationship  (preserves order of elements written/loaded)
	Addresses []*Address
	// manyUser-to-1ChessClub - 2 fields: ChessClub and ChessClubID (mandatory!)
	ChessClub   *ChessClub
	ChessClubID uint
	// manyUser-to-manyAutoBuses (internally uses a join table users_autobuses)
	AutoBuses []*AutoBus `gorm:"many2many:users_autobuses;"` // plural
}

type Profile struct {
	gormCrud.GormCrud[Profile]
	Bio string
	// 1User-to-1Profile relationship (mandatory!)
	UserID uint `gorm:"unique"`
}

type Address struct {
	gormCrud.GormCrud[Address]
	Street  string
	City    string
	State   string
	ZipCode string
	// 1User-to-manyAddress belongs-to relationship (mandatory!)
	UserID uint
}

type ChessClub struct {
	gormCrud.GormCrud[ChessClub]
	Name string
	// manyUser-to-1ChessClub - no fields needed in ChessClub
}

type AutoBus struct {
	gormCrud.GormCrud[AutoBus]
	BusName string
	// manyUser-to-manyAutoBuses (internally uses a join table users_autobuses)
	Users []*User `gorm:"many2many:users_autobuses;"` // plural
}

func main() {

	// Initialize the database with the desired SQLite filename
	{
		var dbfilename string
		if len(os.Args) > 1 {
			dbfilename = os.Args[1]
		} else {
			dbfilename = "/tmp/" + uuid.New().String() + ".sqlite"
		}
		if err := gormCrud.InitializeDB(dbfilename); err != nil {
			panic(err)
		}
	}

	// Migrate models
	models := []interface{}{&User{}, &Profile{}, &Address{}}
	if err := gormCrud.MigrateModels(models...); err != nil {
		panic(err)
	}

	// Create user with new profile and addresses
	u := &User{
		Name:   "Alice",
		Email:  "alice@example.com",
		Age:    30,
		Active: true,
		Profile: &Profile{
			Bio: "Developer",
		},
		Addresses: []*Address{
			{
				Street:  "123 Main St",
				City:    "Wonderland",
				State:   "Imaginary",
				ZipCode: "12345",
			},
			{
				Street:  "456 Side St",
				City:    "Fictionville",
				State:   "Nowhere",
				ZipCode: "67890",
			},
		},
	}

	// Create a new chess club
	x := &ChessClub{Name: "Chess Club"}
	if err := x.Save(x); err != nil {
		panic(err)
	}

	// Set the chess club for the user
	// NOTE: setting u.ChessClubID is enough (no need to also set u.ChessClubID, it will be implicitly updated once u is saved into db)
	u.ChessClub = x

	// Create a new autobus1 and autobus2
	ab1 := &AutoBus{BusName: "Bus1"}
	ab2 := &AutoBus{BusName: "Bus2"}
	// Set the autobuses
	if err := ab1.Save(ab1); err != nil {
		panic(err)
	}
	if err := ab2.Save(ab2); err != nil {
		panic(err)
	}
	// Create a new user2
	u2 := &User{Name: "Bob"}
	// Append the autobuses to the user - u associates with ab1 and ab2
	u.AutoBuses = append(u.AutoBuses, ab1)
	u.AutoBuses = append(u.AutoBuses, ab2)
	// Append the autobus1 to the user2 - u2 associates only with ab1
	u2.AutoBuses = append(u2.AutoBuses, ab1)

	// Save user to database
	if err := u.Save(u); err != nil {
		panic(err)
	}

	// Save user2 to database
	if err := u2.Save(u2); err != nil {
		panic(err)
	}

	// Load from database user
	fmt.Printf("\n\nLoading from db the User records\n")
	{
		results, err := u.LoadWhere("age >= ?", 0)
		if err != nil {
			panic(err)
		}
		printJSON(results)
	}
	fmt.Printf("\n\nLoading from db the AutoBus records\n")
	{
		results, err := (&AutoBus{}).LoadWhere("bus_name != ''", "")
		if err != nil {
			panic(err)
		}
		printJSON(results)
	}

	// manyUser-to-manyAutoBuses
	//
	// User.AutoBuses and AutoBus.Users have a manyUser-to-manyAutoBuses (internally uses a join table users_autobuses)
	// To add an AutoBus into User.AutoBuses[], just append it and save: u.AutoBuses = append(u.AutoBuses, ab1)
	// To remove an AutoBus from User.AutoBuses[], due to the many-to-many relation, we need to do it in a special way,
	// using GORM's Association() method: gormCrud.Db.Model(aUser).Association("AutoBuses").Delete(aAutoBus)
	//

	// Remove ab2 from user u's AutoBuses
	fmt.Printf("\n\nRemoving Bus1 from Alice's auto buses (Alice should end with only Bus2)\n")
	{

		// We are removing ab1 from u.AutoBuses[] in a way that also updates the join table users_autobuses
		if err := gormCrud.Db.Model(u).Association("AutoBuses").Delete(ab1); err != nil {
			panic(err)
		}

		fmt.Printf("\n\nLoading from db the User records (Alice should only be in Bus2)\n")
		{ // Reload the user to get the updated AutoBuses
			if err := u.Reload(u); err != nil {
				panic(err)
			}
			printJSON(u)
		}

		fmt.Printf("\n\nReLoading from db the AutoBus records (Alice should only be in Bus2)\n")
		{
			results, err := (&AutoBus{}).LoadWhere("bus_name != ''", "")
			if err != nil {
				panic(err)
			}
			printJSON(results)
		}
	}
}

func printJSON(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonData))
}
