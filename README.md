# PAASENV
Braindead personal script that takes env vars from a file and sets them for a specified heroku or fly app. Heroku CLI/flyctl is required for this script to work. Should work fine on all systems.
## BUILD
    go build -o whatever-you-want main.go
Build and put the binary in $PATH or bin folder
## USAGE
Providing file name and app name will set the vars

    paasenv --heroku (--fly) -p path/to/file -a app-name

**d** switch means to unset all variables

    paasenv --heroku (--fly) -d -a app-name

**d-keep** switch means to unset all vars and save them in a file, however it is impossible to do it with fly

    paasenv --heroku -d-keep -a app-name

## NOTES
* you have to be logged in Heroku CLI/flyctl
* env var name consist of letters, numbers and _ only
* double quotes at the beginning and end are removed automatically
* if your env var value is supposed to have double quotes at the beginning and end put additional pair of double quotes around
