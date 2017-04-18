package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
		assertNoUnstagedChanges()
		user := findUser(*giveUser)
		message := fmt.Sprintf("git-coin: Giving %v coins to %s", *giveCoins, user)
		fmt.Println(message)
	case donate.FullCommand():
		assertNoUnstagedChanges()
		user := findUser(*donateUser)
		message := fmt.Sprintf("git-coin: Donating %v coins to %s", *giveCoins, user)
		fmt.Println(message)
	case list.FullCommand():
		fmt.Println("TODO: list all coins")
	}
}

func assertNoUnstagedChanges() {
	cmd := exec.Command("git", "diff-index", "--quiet", "HEAD")
	err := cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = cmd.Wait()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			fmt.Fprintln(os.Stderr, "Please stash or commit unstaged changes before exchanging git-coins. Dirty exchanges are unsanitary...")
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(2)
	}
}

func findUser(user string) string {
	// git log -F -i '--author=nathan clark' -1 '--pretty=format:%an <%ae>'
	cmd := exec.Command("git", "log", "-F", "-i", "-1", "--pretty=format:%an <%ae>",
		fmt.Sprintf("--author=%s", user))
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(4)
	}

	actualUser := strings.TrimSpace(string(out))

	if actualUser == "" {
		fmt.Fprintln(os.Stderr, "User", *giveUser, "not found; using", *giveUser, "as-is")
		actualUser = *giveUser
	}

	return actualUser
}
