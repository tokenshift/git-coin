package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/alecthomas/kingpin"
)

var (
	app = kingpin.New("git-coin", "Turn your git repo into a transaction ledger.")

	give      = app.Command("give", "Give coins to another user.")
	giveUser  = give.Arg("user", "The user to give coins to.").Required().String()
	giveCoins = give.Arg("coins", "The number of coins to give.").Required().Float()
	giveForce = give.Flag("force", "Give coins even if it will cause the current user's balance to go negative.").Short('f').Bool()

	take      = app.Command("take", "Take coins from another user.")
	takeUser  = take.Arg("user", "The user to take coins from.").Required().String()
	takeCoins = take.Arg("coins", "The number of coins to take.").Required().Float()
	takeForce = take.Flag("force", "Take coins even if it will cause the target user's balance to go negative.").Short('f').Bool()

	donate      = app.Command("donate", "Add to a user's count of coins (without costing anything). Useful for \"seeding\" the ledger with a certain amount of coins.")
	donateUser  = donate.Arg("user", "The user to donate coins to.").Required().String()
	donateCoins = donate.Arg("coins", "The number of coins to donate.").Required().Float()

	seed      = app.Command("seed", "Give (donate) every user in the commit history a number of starting coins.")
	seedCoins = seed.Arg("coins", "The number of coins to donate to each user.").Required().Float()

	list = app.Command("list", "List all users and how many coins they have.")

	info = app.Command("info", "Tell me how many coins I have.")
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case give.FullCommand():
		assertNoUnstagedChanges()
		user := findUser(*giveUser)

		if *giveCoins > myCoins() && !*giveForce {
			fmt.Fprintln(os.Stderr, "You don't have enough coins!")
			os.Exit(1)
		}

		message := fmt.Sprintf("git-coin: Giving %v coins to %s", *giveCoins, user)
		fmt.Println(message)
		commit(message)
	case take.FullCommand():
		assertNoUnstagedChanges()
		if *takeForce {
			fmt.Println("Seriously, you can't take coins from somebody else. Stop trying.")
		} else {
			fmt.Println("You can't take coins from somebody else. A**hole.")
		}
	case donate.FullCommand():
		assertNoUnstagedChanges()
		user := findUser(*donateUser)
		message := fmt.Sprintf("git-coin: Donating %v coins to %s", *donateCoins, user)
		fmt.Println(message)
		commit(message)
	case seed.FullCommand():
		assertNoUnstagedChanges()
		for _, user := range allUsers() {
			message := fmt.Sprintf("git-coin: Donating %v coins to %s", *seedCoins, user)
			fmt.Println(message)
			commit(message)
		}
	case list.FullCommand():
		listCoins()
	case info.FullCommand():
		user := currentUser()
		coins := myCoins()
		fmt.Printf("You are %s and you have %v coins.\n", user, coins)
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
			fmt.Fprintln(os.Stderr, "Please stash or commit unstaged changes before exchanging git-coins. Dirty transactions are unsanitary...")
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func findUser(user string) string {
	cmd := exec.Command("git", "log", "-F", "-i", "-1", "--pretty=format:%an <%ae>",
		fmt.Sprintf("--author=%s", user))
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	actualUser := strings.TrimSpace(string(out))

	if actualUser == "" {
		fmt.Fprintln(os.Stderr, "User", user, "not found; using", user, "as-is")
		actualUser = user
	}

	return actualUser
}

func commit(message string) {
	cmd := exec.Command("git", "commit", "--allow-empty", "-m", message)
	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func listCoins() {
	ledger := getLedger()
	users := allUsers()

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 4, '.', 0)

	for _, user := range users {
		if coins, ok := ledger[strings.ToUpper(user)]; ok {
			fmt.Fprintf(writer, "%s\t%v\n", user, coins)
		}
	}

	writer.Flush()
}

// `allUsers` returns a list of all users in the git commit history, in order.
// Case-insensitive collisions are omitted.
func allUsers() []string {
	users := make(map[string]string)

	cmd := exec.Command("git", "log", "--pretty=format:%an <%ae>")
	out, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err = cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		user := strings.TrimSpace(scanner.Text())
		norm := strings.ToUpper(user)
		users[norm] = user
	}

	if scanner.Err() != nil {
		fmt.Fprintln(os.Stderr, scanner.Err())
		os.Exit(1)
	}

	if err = cmd.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, scanner.Err())
		os.Exit(1)
	}

	allUsers := make([]string, 0, len(users))
	for _, user := range users {
		allUsers = append(allUsers, user)
	}

	sort.Strings(allUsers)

	return allUsers
}

// `ledger` returns a map of users to their current balances.
// Users are uppercased for consistency.
func getLedger() map[string]float64 {
	ledger := make(map[string]float64)

	cmd := exec.Command("git", "log", "--grep=^git-coin:", "--pretty=format:%an <%ae> - %s")
	out, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err = cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	txPattern := regexp.MustCompile(`(?i)^(.*?) - git-coin: (Giving|Donating) (\d+(\.\d+)?(e(\+|\-)?\d+)?) coins to (.*)$`)

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		matches := txPattern.FindStringSubmatch(scanner.Text())
		if matches == nil {
			continue
		}

		sourceUser := strings.ToUpper(strings.TrimSpace(matches[1]))
		action := strings.TrimSpace(matches[2])
		coins := strings.TrimSpace(matches[3])
		targetUser := strings.ToUpper(strings.TrimSpace(matches[7]))

		coinsNum, err := strconv.ParseFloat(coins, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if _, ok := ledger[sourceUser]; !ok {
			ledger[sourceUser] = 0.0
		}

		if _, ok := ledger[targetUser]; !ok {
			ledger[targetUser] = 0.0
		}

		switch strings.ToUpper(action) {
		case "GIVING":
			ledger[sourceUser] -= coinsNum
			ledger[targetUser] += coinsNum
		case "DONATING":
			ledger[targetUser] += coinsNum
		}
	}

	if scanner.Err() != nil {
		fmt.Fprintln(os.Stderr, scanner.Err())
		os.Exit(1)
	}

	if err = cmd.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, scanner.Err())
		os.Exit(1)
	}

	return ledger
}

func currentUser() string {
	name, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	email, err := exec.Command("git", "config", "user.email").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	nameStr := strings.TrimSpace(string(name))
	emailStr := strings.TrimSpace(string(email))

	if nameStr == "" || emailStr == "" {
		fmt.Fprintln(os.Stderr, "I don't know how much money you have; I don't even know who you are!")

		if nameStr == "" {
			fmt.Fprintln(os.Stderr, "git config user.name")
		}

		if emailStr == "" {
			fmt.Fprintln(os.Stderr, "git config user.email")
		}

		os.Exit(1)
	}

	return fmt.Sprintf("%s <%s>", nameStr, emailStr)
}

func myCoins() float64 {
	user := currentUser()
	ledger := getLedger()

	if coins, ok := ledger[strings.ToUpper(user)]; ok {
		return coins
	} else {
		return 0.0
	}
}
