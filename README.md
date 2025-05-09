# gb (go and build)
gb is a simple build runner, taking commands from environment variables and .env / .gb files

## install
`go install -v github.com/awryme/gb@latest`

## help
```
available env variables:
	- GB: sets build commands
	- GB_RUN: sets run commands
	- GB_SHELL: overrides shell to use

each env for GB and GB_RUN can contain multiple commands
commands are separated by ';'
each command is passed to the shell separately

env variables can be store in .env files
default file is standard '.env', which is always read, if exists
it does not override env variables passed to gb

if an argument is provided, it will be interpreted as an additional env file
only one argument can be specified, if reading it failes, gb prints the error and exits
variables defined in the file override existing env variables
file passed as <name> will be read as <name>, .<name>.gb or .<name>.env, whichever is available in that order

env sources priority:
1) file passed as argument
2) env variables inherited by command
3) .env file
```
