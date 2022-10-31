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
    Dkeep bool      `arg:"--d-keep" help:"delete and keep env vars"`
    Del bool        `arg:"-d,--delete" help:"delete env vars"`
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

func flyPrompt() (bool){
    var response string
    for{
        fmt.Print("Fly variables cannot be saved, continue? (y/n): ")
        fmt.Scanln(&response)
        upperResponse := strings.ToUpper(response)
        if upperResponse == "Y"{
            return true;
        }else if upperResponse == "N"{
            return false;
        }
    }
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

    // name must only consist of letters, digits and _
    isAlNumeric := regexp.MustCompile(`^[a-zA-Z_]+[a-zA-Z0-9_]*$`)
    if !isAlNumeric.MatchString(variableSplit[0]){
        return errors.New("error in variable name")
    }

    return nil
}

func deleteEnvVars()(error){
    if args.Heroku{
        out, err := exec.Command("heroku", "config", "-a", args.App, "--json").CombinedOutput()
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

        unsetCommand := exec.Command("heroku", "config:unset", "-a", args.App)

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
        if args.Dkeep{
            var fileOutputString string
            for key, value := range herokuJSON{
                fileOutputString += fmt.Sprintf("%s=%s\n", key, value)
            }

            fileOutputString = fileOutputString[:len(fileOutputString)-1]

            err = os.WriteFile(generateFilename(), []byte(fileOutputString), 0666)
            if err != nil{
                fmt.Printf("Error saving vars to file: %s\n", err)
            }else{
                fmt.Println("\033[0;34mUnset vars successfully saved!\033[0m")
            }
        }
    }else if args.Fly{
        // get already set vars
        out, err := exec.Command("flyctl", "secrets", "list","-a", args.App, "-j").CombinedOutput()
        if err != nil {
            fmt.Printf("error: getting names of the vars already set: %s\n", err)
            fmt.Printf("fly output: \n%s\n", string(out))
            os.Exit(1)
        }

        var flyJson []struct{
            Name string `json:"Name"`
        }
        err = json.Unmarshal(out, &flyJson)
        if err != nil{
            fmt.Printf("error: unmarshaling name of the variables: %s\n", err)
            os.Exit(1)
        }

        command := exec.Command("flyctl", "secrets", "unset", "--detach", "-a", args.App)
        for _, item := range flyJson{
            command.Args = append(command.Args, item.Name)
        }

        if len(flyJson) == 0{
            fmt.Println("No env vars to unset!")
            os.Exit(0)
        }

        out, err = command.CombinedOutput()
        if err != nil{
            fmt.Printf("error: unsetting vars: %s\n", err)
            fmt.Printf("fly output: \n%s\n", string(out))
            os.Exit(1)
        }else{
            fmt.Println("Unset vars successfully!\n")
        }
    }

    return nil
}

func setVars(variables *string){
    if args.Heroku{
        command := exec.Command("heroku", "config:set", "-a", args.App)

        for _, arg := range strings.Split(*variables, "\n"){
            command.Args = append(command.Args, arg)
        }

        out, err := command.CombinedOutput()
        if err != nil {
            fmt.Printf("error: setting vars: \033[0;31m%s\033[0m\n", err)
            fmt.Printf("Heroku output:\n\033[0;32m%s\033[0m", string(out))
            os.Exit(1)
        }else if strings.Index(string(out), "Enter your Heroku credentials") != -1{
            fmt.Println("\033[0;31mFirst log into heroku cli then run this script!\033[0m")
            os.Exit(1)
        }else if strings.Index(string(out), "and restarting") != -1{
            fmt.Println("\033[0;34mVars successfully set!\033[0m")
        }else{
            fmt.Printf("Heroku output:\n\033[0;32m%s\033[0m", string(out))
            os.Exit(1)
        }
    }else{ // setting fly vars
        command := exec.Command("flyctl", "secrets", "set", "--detach", "-a", args.App)
        for _, arg := range strings.Split(*variables, "\n"){
            command.Args = append(command.Args, arg)
        }

        out, err := command.CombinedOutput()
        if err != nil{
            fmt.Printf("error: settings vars: %s", err)
            fmt.Printf("flyctl output: \n%s", string(out))
            os.Exit(1)
        }else if strings.Index(string(out), "login with") != -1{
            fmt.Printf("first log into flyctl the run this script!")
            os.Exit(1)
        }else if strings.Index(string(out), "Release") != -1{
            fmt.Println("vars successfully set!")
        }else{
            fmt.Printf("flyctl output: \n%s", string(out))
            os.Exit(1)
        }
    }
}

func main(){
    if !(args.Del || args.Dkeep){
        fileContents, err := os.ReadFile(args.Path)
        if err != nil{
            fmt.Printf("error: cannot open file \"%s\"!\n", args.Path)
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
        if (args.Fly && args.Dkeep) && !flyPrompt(){
            os.Exit(0)
        }
        err := deleteEnvVars()
        if err != nil{
            fmt.Printf("error: %s\n", err)
        }
    }
    os.Exit(0)
}
