package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kingpin"
)

var (
	app = kingpin.New("git-coin", "Turn your git repo into a transaction ledger.")

	give      = app.Command("give", "Give coins to another user.")
	giveUser  = give.Arg("user", "The user to give coins to.").Required().String()
	giveCoins = give.Arg("coins", "The number of coins to give.").Required().Float()

	donate      = app.Command("donate", "Add to a user's count of coins (without costing anything). Useful for \"seeding\" the ledger with a certain amount of coins.")
	donateUser  = donate.Arg("user", "The user to give coins to.").Required().String()
	donateCoins = donate.Arg("coins", "The number of coins to give.").Required().Float()

	list = app.Command("list", "List all users and how many coins they have.")
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case give.FullCommand():
		fmt.Println("Giving", *giveUser, *giveCoins, "coins")
	case donate.FullCommand():
		fmt.Println("Donating", *donateCoins, "coins to", *donateUser)
	case list.FullCommand():
		fmt.Println("TODO: list all coins")
	}
}
