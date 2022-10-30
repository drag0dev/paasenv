package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
	arg "github.com/alexflint/go-arg"
)

var args struct{
    Heroku bool     `arg:"-h,--heroku"`
    Fly bool        `arg:"-f,--fly"`
    Dkeep bool      `arg:"--dk,--d-keep" help:"delete and keep env vars"`
    Del bool        `arg:"--d, --delete" help:"delete env vars"`
    App string      `arg:"required,-a,--app" help:"name of the app"`
    Path string     `arg:"required,-p,--path" help:"path to the file with env vars"`
}

func init(){
    arg.MustParse(&args)
    if !(args.Heroku || args.Fly){
        fmt.Println("error: it has to be specified whether you are applying change to fly or heroku!")
        os.Exit(1)
    }
}

func generateFilename() (string) {
    var platform string
    if args.Fly{
        platform = "heroku"
    }else{
        platform = "fly"
    }
    date := time.Now()
    var dateStr = fmt.Sprintf("%d.%d.%d", date.Local().Day(), date.Local().Month(), date.Local().Year())
    var timeStr = fmt.Sprintf("%d.%d.%d", date.Local().Hour(), date.Local().Minute(), date.Local().Second())
    return fmt.Sprintf("%s-%s-%s", platform, dateStr, timeStr)
}

func deleteEnvVars()(error){
    out, err := exec.Command("heroku", "config", "-a", appName, "--json").CombinedOutput()
    if err != nil{
        fmt.Printf("Heroku output:\n\033[0;32m%s\033[0m", string(out))
        return errors.New(fmt.Sprintf("error getting currently set variables: %s", err))
    }

    var herokuJSON map[string]string

    err = json.Unmarshal(out, &herokuJSON)
    if err != nil{
        fmt.Printf("Heroku output:\n\033[0;32m%s\033[0m", string(out))
        return errors.New(fmt.Sprintf("error unmarshaling currently set variables: %s", err))
    }

    unsetCommand := exec.Command("heroku", "config:unset", "-a", appName)

    for key := range herokuJSON{
        unsetCommand.Args = append(unsetCommand.Args, key)
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

    if option == "unset-keep"{
        var fileOutputString string
        for key, value := range herokuJSON{
            fileOutputString += fmt.Sprintf("%s=%s\n", key, value)
        }

        fileOutputString = fileOutputString[:len(fileOutputString)-1]

        err = os.WriteFile(saveFileName, []byte(fileOutputString), 0666)
        if err != nil{
            fmt.Printf("Error saving vars to file: %s\n", err)
        }else{
            fmt.Println("\033[0;34mUnset vars successfully saved!\033[0m")
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
    isAlNumeric := regexp.MustCompile(`^[a-zA-Z_]+[a-zA-Z0-9_]*$`)
    if !isAlNumeric.MatchString(variableSplit[0]){
        return errors.New("error in variable name")
    }

    return nil
}

func setVars(variables *string){
    command := exec.Command("heroku", "config:set", "-a", appName)

    for _, arg := range strings.Split(*variables, "\n"){
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
    if !(args.Del && args.Dkeep){
        fileContents, err := os.ReadFile(args.Path)
        if err != nil{
            fmt.Printf("error: cannot open file \"%s\"!\n", fileName)
            os.Exit(1)
        }

        fileContentsStr := string(fileContents)
        if len(fileContents) == 0 {
            fmt.Println("error: file empty")
            os.Exit(1)
        }
        fileContentsStr = strings.TrimSpace(fileContentsStr)

        // check if all vars are valid
        for index, envVar := range strings.Split(fileContentsStr, "\n"){
            err := checkVar(&envVar)
            if err != nil{
                fmt.Printf("error: %s (var number: %d)\n", err, index+1)
                os.Exit(1)
            }
        }
        setVars(&fileContentsStr)
    }else{
        err := deleteEnvVars()
        if err != nil{
            fmt.Printf("error: %s\n", err)
        }
    }
    os.Exit(0)
}
