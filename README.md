#Minibank
A very simple rest api service to teach about cloud computing, containerization and distributed systems.


## Description 

The following are the supported user operations

1) register account
2) authenticate account
3) make deposit
4) schedule payment to other accounts
5) check balance
6) get statement


## Account registration

To register an account you will need to provide a username and password. The following validations are done:

- the username must not already exist
- the username must be alphanumeric
- the password must be at least 10 characters long

## make deposit

The make deposit account requires:

- an amount
- an account where deposits are transfered from
- a document identifier

## schedule payment

To schedule a payment:

- the id of the payee
- the amount
- the posting date

## check balance

This returns the current balance. Accepts the following query parameter:

- asof: returns the balance as of the end of day of the specific date. If the date is before the account opening date, it returns an error.

## get statetent

Returns the transactions in the current period.  Periods are based on the first day of the month. Accepts the following query parameter:

- month and year: returns the list of transactions for the specified period 


