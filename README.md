# daytrading
SENG 468 Project

TODO Group

## Performance
Our system that has been optimized for performance is currently on our master branch. This is the branch that we used to run the final workload.

## Testing
The test suite is on the tests branch within the Servers/transaction_server folder. It is named redis_commands_test.go. 

To run the test suite:
1. open a redis server at port 6380 using the command `redis-server --port 6380`
2. build and run the mock quote server using the commands `go build` and `./quote_server` inside Servers/quote_server
3. go to Servers/transaction_server and run `go test`

## UI and error handling
The version of our system that has been optimized for the user interface with error handling is located on the sep_error branch. Error handling is done within the transaction server while the web server handles the user interface.
