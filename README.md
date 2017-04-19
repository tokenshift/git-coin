# git-coin

_git commit -m "commerce"_

Turn your git repo into a bank!

## Overview

Treats your git repo like it's a transaction log, with each commit as a transaction. Use
the `seed` command to give everyone some starting funds (who says wealth creation is hard),
then `give` your coins to other users

## Installation

```
go get -u github.com/tokenshift/git-coin
go install github.com/tokenshift/git-coin
```

## Usage

```
# Give everyone in the git commit history 10 coins
$ git coin seed 10
Giving 10 coins to George Foobar <george.foobar@example.com>
Giving 10 coins to Melissa Fizzbuzz <mfizzbuz@example.com>

# Give 6 of your coins to "george"
$ git coin give george 6
Giving 6 coins to George Foobar <george.foobar@example.com>

# Check how many coins you have
$ git coin info
You are Melissa Fizzbuzz <mfizzbuz@example.com> and you have 4 coins.

# See how many coins everyone else has
$ git coin list
George Foobar <george.foobar@example.com>....16
Melissa Fizzbuzz <mfizzbuz@example.com>......4
```

For full help:

```
git coin help
```

## Disclaimer

Gitcoins are not real money, and cannot be exchanged for real money.
