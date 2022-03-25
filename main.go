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
var saveFileName string
var appName string
var option string

func init(){
    // get file name from args
    if len(os.Args[1:]) <= 1 {
        fmt.Println("Error: missing arguments!")
        os.Exit(0)
    }

    if os.Args[1] == "-d"{
        // unsetting
        option = "unset"
        if len(os.Args[2]) == 0{
            fmt.Println("Error: app name is missing")
        }
        appName = os.Args[2]
    }else if os.Args[1] == "-dk" && len(os.Args) == 4{
        // unsetting and saving
        if len(os.Args[2]) == 0{
            fmt.Println("Error: missing save file name!")
            os.Exit(0)
        }
        saveFileName = os.Args[2]
        if len(os.Args[3]) == 0{
            fmt.Println("Error: app name is missing")
        }
        appName = os.Args[3]
        option = "unset-keep"
    }else{
        // setting vars
        if len(os.Args[2])==0{
            fmt.Println("Error: env file is missing")
            os.Exit(0)
        }
        fileName = os.Args[1]
        appName = os.Args[2]
        option = "set"
    }
}

func deleteEnvVars()(error){
    out, err := exec.Command("heroku", "config", "-a", appName).CombinedOutput()
    if err != nil{
        fmt.Printf("Heroku output:\n\033[0;32m%s\033[0m", string(out))
        return errors.New(fmt.Sprintf("error getting currently set variables: %s", err))
    }

    herokuOutput := strings.Split(string(out), "\n")

    // remove the first info line
    herokuOutput = herokuOutput[1:]

    // remove the last empty line
    if len(herokuOutput[len(herokuOutput)-1]) == 0{
        herokuOutput = herokuOutput[:len(herokuOutput)-1]
    }

    unsetCommand := exec.Command("heroku", "config:unset", "-a", appName)

    for _, variable := range herokuOutput{
        unsetCommand.Args = append(unsetCommand.Args, strings.Split(variable, ":")[0])
    }

    out, err = unsetCommand.CombinedOutput()
    if err != nil{
        fmt.Printf("Heroku output:\n\033[0;32m%s\033[0m", string(out))
        return errors.New(fmt.Sprintf("error unsetting variables: %s", err))
    }else if strings.Index(string(out), "done") != -1{
        fmt.Println("\033[0;34mVars sucessfully unset!\033[0m")
    }else if strings.Index(string(out), "Enter your Heroku credentials") != -1{
        fmt.Println("\033[0;31mFirst log into heroku cli then run this script!\033[0m")
    }else{
        fmt.Printf("Heroku output:\n\033[0;32m%s\033[0m", string(out))
    }

    // TODO: when setting vars there cannot be spaces between value and =
    if option == "unset-keep"{
        err = os.WriteFile(saveFileName, []byte(strings.ReplaceAll(strings.Join(herokuOutput, "\n"), ":", "=")), 0666)
        if err != nil{
            fmt.Printf("Error saving vars to file: %s\n", err)
        }else{
            fmt.Println("\033[0;34mVars sucessfully unset!\033[0m")
        }
    }

    return nil
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
    // TODO: heroku cli funny when there are spaces in var value
    *variables = strings.ReplaceAll(*variables, "\n", " ")
    *variables = strings.ReplaceAll(*variables, "\"", "")

    command := exec.Command("heroku", "config:set", "-a", appName)

    for _, arg := range strings.Split(*variables, " "){
        command.Args = append(command.Args, arg)
    }

    out, err := command.CombinedOutput()
    if err != nil {
        fmt.Printf("Error: error setting vars: \033[0;31m%s\033[0m\n", err)
        fmt.Printf("Heroku output:\n\033[0;32m%s\033[0m", string(out))
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
    if option == "set"{
        fileContents, err := os.ReadFile(fileName)
        if err != nil{
            fmt.Printf("Error: cannot open file \"%s\"!\n", fileName)
            os.Exit(0)
        }

        fileContentsStr := string(fileContents)
        if len(fileContents) == 0 {
            fmt.Println("Error: file empty")
            os.Exit(0)
        }
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
    }else {
        err := deleteEnvVars()
        if err != nil{
            fmt.Printf("Error: %s\n", err)
        }
    }
    os.Exit(0)
}
