package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var fileName string
var appName string

func init(){
    // TODO: "-d" delete all
    // get file name from args
    if len(os.Args[1:]) <= 1 {
        fmt.Println("Error: missing arguments!")
        os.Exit(0)
    }
    fileName = os.Args[1]
    appName = os.Args[2]
}

func checkVar(variable *string)(error){
    var variableSplit []string = strings.Split(*variable, "=")
    if len(variableSplit) != 2 {
        return errors.New("missing parts of the variable")
    }else if len(variableSplit[0]) == 0 {
        return errors.New("missing variable name")
    }else if len(variableSplit[1]) == 0 {
        return errors.New("missing variable value")
    }

    // name must only consist of uppercase letters, digits and _
    isAlNumeric := regexp.MustCompile(`[a-zA-Z_]+[a-zA-Z0-9_]*`)
    if !isAlNumeric.MatchString(variableSplit[0]){
        return errors.New("error in variable name")
    }

    return nil
}

func setVars(variables *string){
    *variables = strings.ReplaceAll(*variables, "\n", " ")
    *variables = strings.ReplaceAll(*variables, "\"", "")

    command := exec.Command("heroku", "config:set", "-a", appName)

    for _, arg := range strings.Split(*variables, " "){
        command.Args = append(command.Args, arg)
    }

    out, err := command.CombinedOutput()
    if err != nil {
        fmt.Printf("Error: error setting vars: \033[0;31m%s\033[0m\n", err)
        os.Exit(0)
    }else if strings.Index(string(out), "Enter your Heroku credentials") != -1{
        fmt.Println("\033[0;31mFirst log into heroku cli then run this script!\033[0m")
    }else if strings.Index(string(out), "and restarting") != -1{
        fmt.Println("\033[0;34mVars sucessfully set!\033[0m")
    }else{
        fmt.Printf("Heroku output:\n\033[0;32m%s\033[0m", string(out))
    }
}

func main(){
    fileContents, err := os.ReadFile(fileName)
    if err != nil{
        fmt.Printf(`Error: cannot open file "%s"!\n`, fileName)
    }

    fileContentsStr := string(fileContents)
    if fileContentsStr[len(fileContentsStr)-1] == '\n'{
        fileContentsStr = fileContentsStr[:len(fileContentsStr)-1]
    }

    // check if all vars are valid
    for index, envVar := range strings.Split(fileContentsStr, "\n"){
        err := checkVar(&envVar)
        if err != nil{
            fmt.Printf("Error: %s (var number: %d)\n", err, index+1)
            os.Exit(0)
        }
    }

    // set vars
    setVars(&fileContentsStr)
    os.Exit(0)
}
