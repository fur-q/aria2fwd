package main

import (
    "bufio"
    "bytes"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "os/exec"
    "os/user"
    "path/filepath"
    "github.com/kballard/go-shellquote"
)

var (
    cfg arconfig
)

type arconfig struct {
    Addr, Success, Error string
}

type jr_req struct {
    Jsonrpc string `json:"jsonrpc"`
    Id      string `json:"id"`
    Method  string `json:"method"`
    Params  []interface{} `json:"params"`
}

type jr_err struct {
    Jsonrpc string `json:"jsonrpc"`
    Id string `json:"id"`
    Error struct {
        Code int `json:"code"`
        Message string `json:"message"`
    } `json:"error"`
}

type jr_ok struct {
    Jsonrpc string `json:"jsonrpc"`
    Id      string `json:"id"`
    Result  string `json:"result"`
}

func configPath() string {
    usr, err := user.Current()
    if err != nil { panic(err) }
    return filepath.Join(usr.HomeDir, ".aria2fwd")
}

func prompt(rdr *bufio.Reader, str, app string) string {
    if app == "" {
        app = "none set"
    }
    fmt.Printf(str, app)
    out, _, err := rdr.ReadLine()
    if err != nil { panic(err) }
    if out == nil { return app }
    return string(out)
}

func readConfig() (error) {
    fcon, err := ioutil.ReadFile(configPath())
    if err != nil {
        return err
    }
    return json.Unmarshal(fcon, &cfg)
}

// TODO: default config values
func writeConfig() error {
    stdin := bufio.NewReader(os.Stdin)
    fmt.Println("Creating new config file...")

    cfg.Addr    = prompt(stdin, "aria2 server address (%s): ", cfg.Addr)
    cfg.Success = prompt(stdin, "Command to run on success (%s): ", cfg.Success)
    cfg.Error   = prompt(stdin, "Command to run on error (%s): ", cfg.Error)

    out, err := json.Marshal(cfg)
    if err != nil {
        return err
    }
    err = ioutil.WriteFile(filepath.Join(configPath()), out, 0644)
    if err == nil {
        err = errors.New("Successfully wrote config.json")
    }
    return err
}

func addUri(uri string) ([]byte, error) {
    params := []interface{}{ []string{ uri } }
    return json.Marshal(jr_req { "2.0", "a", "aria2.addUri", params })
}

func addTorrent(path string) ([]byte, error) {
    fcon, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }
    params := []interface{} { base64.StdEncoding.EncodeToString(fcon) }
    return json.Marshal(jr_req { "2.0", "a", "aria2.addTorrent", params })
}

func postRequest(req []byte, addr string) error {
    var cmd string
    rdr := bytes.NewReader(req)
    add := fmt.Sprintf("http://%s/jsonrpc", addr)
    res, err := http.Post(add, "application/x-www-form-urlencoded", rdr)
    if err != nil { return err }
    rbd, err := ioutil.ReadAll(res.Body)
    if res.StatusCode == 200 {
        var jres jr_ok
        if cfg.Success == "" { return nil }
        json.Unmarshal(rbd, &jres)
        cmd = fmt.Sprintf(cfg.Success, jres.Result)
    } else {
        var jres jr_err
        if cfg.Error == "" { return nil }
        json.Unmarshal(rbd, &jres)
        cmd = fmt.Sprintf(cfg.Error, jres.Error.Message)
    }
    arg, err := shellquote.Split(cmd)
    if err != nil { return err }
    run := exec.Command(arg[0], arg[1:]...)
    return run.Start()
}

func main() {
    var req []byte
    var err error

    if len(os.Args) <= 1 {
        fmt.Println("usage: aria2fwd (-c) URL|MAGNET-URI|TORRENT-PATH")
        return
    }
    arg := os.Args[1]
    cfg, err = readConfig()
    if err != nil || cfg.Addr == "" {
        if !os.IsNotExist(err) { goto Error }
        fmt.Println("No config file found.")
        err = writeConfig()
        if err != nil { goto Error }
    }
    switch os.Args[1] {
        case "-c":
            err = writeConfig()
            if err != nil { break }
            if len(os.Args) > 1 {
                arg = os.Args[2]
                fallthrough
            }
        default:
            req, err = addTorrent(os.Args[1])
            // FIXME check err here or rework this
            //if os.IsNotExist(err) {
            if err != nil {
                req, err = addUri(os.Args[1])
            }
    }
    if err != nil { goto Error }
    err = postRequest(req, cfg.Addr)
    if err != nil { goto Error }

Error:
    if err != nil {
        fmt.Println(err)
    }
}
