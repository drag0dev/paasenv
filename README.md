# HRKVAR
Script that takes env vars from a file and sets them for a specified heroku app. Heroku CLI is required for this script to work. Should work fine on all systems.

## BUILD
    go build -o hrkvar main.go
Build and put the binary in $PATH or bin folder
## USAGE
Providing file name and app name will set the vars

    hrkvar path/to/file app-name

d flag means to unset all variables

    hrkvar -d app-name

dk flag means to unset all vars and save (keep) them in a file

    hrkvar -dk path/to/save-file app-name

## NOTES
* you have to be logged in Heroku CLI
* env var name consist of letters, numbers and _ only
* double quotes at the beginning and end are removed automatically
* if your env var value is supposed to have double quotes at the beginning and end put additional pair of double quotes around
