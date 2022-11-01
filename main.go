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
  color "github.com/fatih/color"
)

var red func (a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var yellow func (a ...interface{}) string = color.New(color.FgYellow).SprintFunc()
var green func (a ...interface{}) string = color.New(color.FgGreen).SprintFunc()

var args struct{
    Heroku bool     `arg:"--heroku"`
    Fly bool        `arg:"-f,--fly"`
    Dkeep bool      `arg:"--d-keep" help:"delete and keep env vars"`
    Del bool        `arg:"-d,--delete" help:"delete env vars"`
    App string      `arg:"required,-a,--app" help:"name of the app"`
    Path string     `arg:"-p,--path" help:"path to the file with env vars"`
}

func init(){
    arg.MustParse(&args)
    if !(args.Heroku || args.Fly){
        fmt.Printf("%s: it has to be specified whether you are applying changes to fly or heroku!\n", red("error"))
        os.Exit(1)
    }
    if !(args.Del || args.Dkeep) && len(args.Path) == 0{
        fmt.Printf("%s: path to env vars is required!\n", red("error"))
        os.Exit(1)
    }
}

func generateFilename() (string) {
    var platform string
    if args.Fly{
        platform = "fly"
    }else{
        platform = "heroku"
    }
    date := time.Now()
    var dateStr = fmt.Sprintf("%d.%d.%d", date.Local().Day(), date.Local().Month(), date.Local().Year())
    var timeStr = fmt.Sprintf("%d.%d.%d", date.Local().Hour(), date.Local().Minute(), date.Local().Second())
    return fmt.Sprintf("%s-%s-%s", platform, dateStr, timeStr)
}

func flyPrompt() (bool){
    var response string
    for{
        fmt.Print("fly variables cannot be saved, continue? (y/n): ")
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
        return errors.New("variable name")
    }

    return nil
}

func deleteEnvVars(){
    if args.Heroku{
        out, err := exec.Command("heroku", "config", "-a", args.App, "--json").CombinedOutput()
        if err != nil{
            fmt.Printf("%s: getting currently set variables: %s\n", red("error"), err)
            fmt.Printf("heroku output:\n%s\n", yellow(string(out)))
            os.Exit(1)
        }

        var herokuJSON map[string]string

        err = json.Unmarshal(out, &herokuJSON)
        if err != nil{
            fmt.Printf("%s: unmarshalling currently set variables: %s\n", red("error"), err)
            fmt.Printf("heroku output:\n%s\n", yellow(string(out)))
            os.Exit(1)
        }

        unsetCommand := exec.Command("heroku", "config:unset", "-a", args.App)

        for key := range herokuJSON{
            unsetCommand.Args = append(unsetCommand.Args, key)
        }

        out, err = unsetCommand.CombinedOutput()
        if err != nil{
            fmt.Printf("%s: unsetting variables: %s\n", red("error"), err)
            fmt.Printf("heroku output:\n%s\n", yellow(string(out)))
            os.Exit(1)
        }else if strings.Index(string(out), "done") != -1{
            fmt.Println(green("vars successfully unset!"))
        }else if strings.Index(string(out), "login") != -1{
            fmt.Println(red("first log into heroku cli then run this script!"))
            os.Exit(1)
        }else{
            fmt.Printf("heroku output:\n%s\n", yellow(string(out)))
            os.Exit(1)
        }
        if args.Dkeep{
            var fileOutputString string
            for key, value := range herokuJSON{
                fileOutputString += fmt.Sprintf("%s=%s\n", key, value)
            }

            fileOutputString = fileOutputString[:len(fileOutputString)-1]

            err = os.WriteFile(generateFilename(), []byte(fileOutputString), 0666)
            if err != nil{
                fmt.Printf("%s: saving vars to file: %s\n", red("error"), err)
            }else{
                fmt.Println(green("unset vars successfully saved!"))
            }
        }
    }else if args.Fly{
        // get already set vars
        out, err := exec.Command("flyctl", "secrets", "list","-a", args.App, "-j").CombinedOutput()
        if err != nil {
            fmt.Printf("%s: getting names of the vars already set: %s\n", red("error"), err)
            fmt.Printf("fly output: \n%s\n", yellow(string((out))))
            os.Exit(1)
        }

        var flyJson []struct{
            Name string `json:"Name"`
        }
        err = json.Unmarshal(out, &flyJson)
        if err != nil{
            fmt.Printf("%s: unmarshalling name of the variables: %s\n", red("error"), err)
            os.Exit(1)
        }

        command := exec.Command("flyctl", "secrets", "unset", "--detach", "-a", args.App)
        for _, item := range flyJson{
            command.Args = append(command.Args, item.Name)
        }

        if len(flyJson) == 0{
            fmt.Println(yellow("no env vars to unset!"))
            os.Exit(0)
        }

        out, err = command.CombinedOutput()
        if err != nil{
            fmt.Printf("%s: unsetting vars: %s\n", red("error"), err)
            fmt.Printf("flyctl output: \n%s", yellow(string(out)))
            os.Exit(1)
        }else if strings.Index(string(out), "access token") != -1{
            fmt.Println(red("first log into flyctl the run this script!"))
            os.Exit(1)
        }else if strings.Index(string(out), "Release") != -1{
            fmt.Println(green("unset vars successfully!"))
        }else{
            fmt.Printf("flyctl output: \n%s\n", yellow(string(out)))
            os.Exit(1)
        }
    }
}

func setVars(variables *string){
    if args.Heroku{
        command := exec.Command("heroku", "config:set", "-a", args.App)

        for _, arg := range strings.Split(*variables, "\n"){
            command.Args = append(command.Args, arg)
        }

        out, err := command.CombinedOutput()
        if err != nil {
            fmt.Printf("%s: setting vars: %s\n", red("error"), yellow(err))
            fmt.Printf("heroku output: \n%s\n", yellow(string(out)))
            os.Exit(1)
        }else if strings.Index(string(out), "login") != -1{
            fmt.Println(red("first log into heroku cli then run this script!"))
            os.Exit(1)
        }else if strings.Index(string(out), "and restarting") != -1{
            fmt.Println(green("vars successfully set!"))
        }else{
            fmt.Printf("heroku output:\n%s\n", yellow(string(out)))
            os.Exit(1)
        }
    }else{ // setting fly vars
        command := exec.Command("flyctl", "secrets", "set", "--detach", "-a", args.App)
        for _, arg := range strings.Split(*variables, "\n"){
            command.Args = append(command.Args, arg)
        }

        out, err := command.CombinedOutput()
        if err != nil{
            fmt.Printf("%s: settings vars: %s\n", red("error"), err)
            fmt.Printf("flyctl output: \n%s", yellow(string(out)))
            os.Exit(1)
        }else if strings.Index(string(out), "access token") != -1{
            fmt.Println(red("first log into flyctl the run this script!"))
            os.Exit(1)
        }else if strings.Index(string(out), "Release") != -1{
            fmt.Println(green("vars successfully set!"))
        }else{
            fmt.Printf("flyctl output: \n%s\n", yellow(string(out)))
            os.Exit(1)
        }
    }
}

func main(){
    fmt.Println(generateFilename())
    if !(args.Del || args.Dkeep){
        fileContents, err := os.ReadFile(args.Path)
        if err != nil{
            fmt.Printf("%s: cannot open file \"%s\"!\n", red("error"), args.Path)
            os.Exit(1)
        }

        fileContentsStr := string(fileContents)
        if len(fileContents) == 0 {
            fmt.Printf("%s: file empty\n", red("error"))
            os.Exit(1)
        }
        fileContentsStr = strings.TrimSpace(fileContentsStr)

        // check if all vars are valid
        for index, envVar := range strings.Split(fileContentsStr, "\n"){
            err := checkVar(&envVar)
            if err != nil{
                fmt.Printf("%s: %s (var number: %d)\n", red("error"), err, index+1)
                os.Exit(1)
            }
        }
        setVars(&fileContentsStr)
    }else{
        if (args.Fly && args.Dkeep) && !flyPrompt(){
            os.Exit(0)
        }
        deleteEnvVars()
    }
    os.Exit(0)
}
