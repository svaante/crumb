package crumb

import (
    "io/ioutil"
    "os"
    "os/exec"
    "log"
    "fmt"
)

func preferedEditor() string {
    editor := os.Getenv("EDITOR")
    if editor == "" {
        return "vi"

    }
    executable, err := exec.LookPath(editor)
    if err != nil {
        log.Fatal(fmt.Sprintf("Could not resolve $EDITOR executable, %s", editor))
    }
    return executable
}

func editWithEditor(original string) string {
    f, err := ioutil.TempFile(os.TempDir(), "*")
    if err != nil {
        log.Fatal(fmt.Sprintf("Could not create tmp file"))
    }

    if _, err = f.Write([]byte(original)); err != nil {
        log.Fatal(fmt.Sprintf("Culd not write to tmp file"))
    }

    filename := f.Name()

    defer f.Close()
    defer os.Remove(filename)

    editor := preferedEditor()

    cmd := exec.Command(editor, filename)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    err = cmd.Start()
    if err != nil {
        log.Fatal(fmt.Sprintf("Could not start %s", editor))
    }
    err = cmd.Wait()

    if err != nil {
        log.Fatal(fmt.Sprintf("Editor %s exited with a non-zero status", editor))
    }

    return readFile(filename)
}
