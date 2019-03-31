# daytrading
SENG 468 Project

transaction_server error cases:

Add:
1. If amount is <0, then show error message and return message to web serve

Buy:
1. If user doesn’t’t exist, add user into db with balance 0
2. If user doesn’t have balance to buy the stock, log error message and return message to web server
3. If stock to buy is 0(unit stock price > amount), log error message and return message to web serve

Commit buy:
1. If balance < totalCost, log error message and return message to web serve
2. If no buy to commit, log error message and return message to web serve

Cancel buy:
1. If no buy to cancel, log error message and return message to web serve

sellHandler
1. If user doesn’t’t exist, add user into db with balance 0
2. If user doesn’t have enough stock to sell required amount, log error message and return message to web server
3. If stock to sell is 0(unit stock price > amount), log error message and return message to web serve

Commit sell:
1. If stockOwned < stockNeeded, log error message and return
2. If no sell to commit, log error message and return message to web serve

Cancel_sell:
1. If no sell to cancel, log error message and return message to web serve
